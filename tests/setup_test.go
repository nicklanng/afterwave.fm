package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"log/slog"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/sopatech/afterwave.fm/internal/artists"
	"github.com/sopatech/afterwave.fm/internal/auth"
	"github.com/sopatech/afterwave.fm/internal/cognito"
	"github.com/sopatech/afterwave.fm/internal/feed"
	"github.com/sopatech/afterwave.fm/internal/follows"
	apphttp "github.com/sopatech/afterwave.fm/internal/http"
	"github.com/sopatech/afterwave.fm/internal/infra"
	"github.com/sopatech/afterwave.fm/internal/metrics"
	"github.com/sopatech/afterwave.fm/internal/search"
	"github.com/sopatech/afterwave.fm/internal/users"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	testTable             string
	testDB                *infra.Dynamo
	testJWTPrivKey        *rsa.PrivateKey
	testJWTPubKey         *rsa.PublicKey
	testOpenSearchEndpoint string
	testFeedIndexName     string
)

// fakeCognitoClient is an in-memory implementation of cognito.Client for tests.
type fakeCognitoClient struct {
	mu     sync.Mutex
	users  map[string]struct {
		password string
		sub      string
	}
}

var _ cognito.Client = (*fakeCognitoClient)(nil)

func newFakeCognitoClient() *fakeCognitoClient {
	return &fakeCognitoClient{
		users: make(map[string]struct {
			password string
			sub      string
		}),
	}
}

func (f *fakeCognitoClient) SignUp(ctx context.Context, email, password string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.users[email]; exists {
		return "", errors.New("user already exists")
	}
	sub := email + "-sub"
	f.users[email] = struct {
		password string
		sub      string
	}{password: password, sub: sub}
	return sub, nil
}

func (f *fakeCognitoClient) InitiateAuth(ctx context.Context, email, password string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[email]
	if !ok || u.password != password {
		return "", errors.New("invalid credentials")
	}
	return u.sub, nil
}

func (f *fakeCognitoClient) AdminDeleteUser(ctx context.Context, email string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.users, email)
	return nil
}

func TestMain(m *testing.M) {
	testTable = getEnv("DYNAMO_TABLE", "afterwave-test")
	region := getEnv("AWS_REGION", "us-east-1")
	endpoint := getEnv("DYNAMODB_ENDPOINT", "http://localhost:8001")

	ctx := context.Background()
	db, err := infra.NewDynamo(ctx, region, endpoint)
	if err != nil {
		slog.Default().Error("dynamo init", "err", err)
		os.Exit(1)
	}
	testDB = db

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		slog.Default().Error("generate test RSA key", "err", err)
		os.Exit(1)
	}
	testJWTPrivKey = key
	testJWTPubKey = &key.PublicKey

	if err := ensureTable(ctx, db, testTable); err != nil {
		slog.Default().Error("ensure table", "err", err)
		os.Exit(1)
	}

	// Clear DynamoDB and OpenSearch between test runs so each run starts clean.
	if err := clearDynamoTable(ctx, db, testTable); err != nil {
		slog.Default().Error("clear dynamo table", "err", err)
		os.Exit(1)
	}
	testOpenSearchEndpoint = getEnv("OPENSEARCH_ENDPOINT", "http://localhost:9200")
	testFeedIndexName = getEnv("OPENSEARCH_FEED_INDEX", "afterwave-feed")
	if testOpenSearchEndpoint != "" {
		if err := clearOpenSearchFeedIndex(ctx, testOpenSearchEndpoint, testFeedIndexName); err != nil {
			slog.Default().Warn("clear opensearch feed index (skipping if OpenSearch not running)", "err", err)
		}
	}

	code := m.Run()
	os.Exit(code)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ensureTable(ctx context.Context, db *infra.Dynamo, table string) error {
	client := db.Client()
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	})
	if err == nil {
		return nil
	}
	var notFound *types.ResourceNotFoundException
	if !errors.As(err, &notFound) {
		return err
	}
	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(table),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("sk"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("sk"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	return err
}

// clearDynamoTable deletes all items in the table (scan then batch delete).
func clearDynamoTable(ctx context.Context, db *infra.Dynamo, table string) error {
	client := db.Client()
	var lastKey map[string]types.AttributeValue
	for {
		scanOut, err := client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(table),
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			return err
		}
		if len(scanOut.Items) == 0 {
			if scanOut.LastEvaluatedKey == nil {
				break
			}
			lastKey = scanOut.LastEvaluatedKey
			continue
		}
		const batchSize = 25
		for i := 0; i < len(scanOut.Items); i += batchSize {
			end := i + batchSize
			if end > len(scanOut.Items) {
				end = len(scanOut.Items)
			}
			chunk := scanOut.Items[i:end]
			writeReqs := make([]types.WriteRequest, 0, len(chunk))
			for _, item := range chunk {
				pk, ok := item["pk"]
				if !ok {
					continue
				}
				sk, ok := item["sk"]
				if !ok {
					continue
				}
				writeReqs = append(writeReqs, types.WriteRequest{
					DeleteRequest: &types.DeleteRequest{
						Key: map[string]types.AttributeValue{"pk": pk, "sk": sk},
					},
				})
			}
			if len(writeReqs) == 0 {
				continue
			}
			_, err = client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{table: writeReqs},
			})
			if err != nil {
				return err
			}
		}
		lastKey = scanOut.LastEvaluatedKey
		if lastKey == nil {
			break
		}
	}
	return nil
}

// clearOpenSearchFeedIndex deletes the feed index so the next run starts with an empty index.
func clearOpenSearchFeedIndex(ctx context.Context, baseURL, indexName string) error {
	osClient := infra.NewOpenSearch(baseURL, nil)
	idx := search.NewFeedIndex(osClient, indexName)
	return idx.DeleteIndex(ctx)
}

// Test auth clients (public, no secret). TTLs: web 15min/7d, native 30d/90d.
const (
	testWebSessionSec, testWebRefreshSec      = 15 * 60, 7 * 24 * 3600
	testNativeSessionSec, testNativeRefreshSec = 30 * 24 * 3600, 90 * 24 * 3600
)

// newTestServer builds the app router and returns an httptest.Server and base URL for the v1 API (e.g. http://example.com/v1).
func newTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	logger := slog.Default()
	ctx := context.Background()

	authStore := auth.NewStore(testDB, testTable)
	testAuthClients := []auth.ClientCredential{
		{ID: "web", SessionTTLSeconds: testWebSessionSec, RefreshTTLSeconds: testWebRefreshSec},
		{ID: "ios", SessionTTLSeconds: testNativeSessionSec, RefreshTTLSeconds: testNativeRefreshSec},
		{ID: "android", SessionTTLSeconds: testNativeSessionSec, RefreshTTLSeconds: testNativeRefreshSec},
		{ID: "desktop", SessionTTLSeconds: testNativeSessionSec, RefreshTTLSeconds: testNativeRefreshSec},
	}
	if err := authStore.EnsureAuthClients(ctx, testAuthClients); err != nil {
		t.Fatalf("ensure auth clients: %v", err)
	}
	mauStore := metrics.NewMAUStore(testDB, testTable)
	metricsReg := prometheus.NewRegistry()
	mauRecorder, err := metrics.NewMAURecorder(mauStore, metricsReg)
	if err != nil {
		t.Fatalf("new MAU recorder: %v", err)
	}
	authSvc := auth.NewService(authStore, testJWTPrivKey, mauRecorder)
	cookieCfg := auth.CookieConfig{Secure: false} // HTTP in tests
	authH := auth.NewHandler(authSvc, cookieCfg)

	userStore := users.NewStore(testDB, testTable)
	userSvc := users.NewService(userStore, newFakeCognitoClient())
	// For tests we don't exercise federated endpoints; pass empty Cognito Hosted UI config.
	userH := users.NewHandler(userSvc, authSvc, cookieCfg, "", "", "", "", "", "", "", "")

	artistStore := artists.NewStore(testDB, testTable)
	artistMemberStore := artists.NewMemberStore(testDB, testTable)
	artistSvc := artists.NewService(artistStore, artistMemberStore)
	artistH := artists.NewHandler(artistSvc)

	followsStore := follows.NewStore(testDB, testTable)
	followsSvc := follows.NewService(followsStore, artistSvc)
	followsH := follows.NewHandler(followsSvc)

	feedStore := feed.NewStore(testDB, testTable)
	var feedSvc feed.Service
	if testOpenSearchEndpoint != "" {
		osClient := infra.NewOpenSearch(testOpenSearchEndpoint, nil)
		feedIndex := search.NewFeedIndex(osClient, testFeedIndexName)
		if err := feedIndex.EnsureIndex(ctx); err != nil {
			t.Logf("opensearch ensure index (my feed tests may be skipped): %v", err)
		}
		feedSvc = feed.NewServiceWithSearch(feedStore, artistSvc, artistSvc, feedIndex, followsSvc, feedIndex)
	} else {
		feedSvc = feed.NewService(feedStore, artistSvc)
	}
	feedH := feed.NewHandler(feedSvc)

	handler := apphttp.NewRouter(logger, userH, authH, artistH, followsH, feedH, metrics.HandlerForRegistry(metricsReg), testJWTPubKey)
	server := httptest.NewServer(handler)
	base := server.URL + "/v1"
	return server, base
}
