package infra

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

// OpenSearch is a client for talking to an OpenSearch (or Elasticsearch) cluster.
// Create via NewOpenSearch; pass to search packages that need index access.
// DoWithRetry performs HTTP requests with retry on transient connection/timeout errors.
type OpenSearch struct {
	BaseURL string
	Client  *http.Client
}

// NewOpenSearch returns an OpenSearch client for the given endpoint (e.g. http://localhost:9200).
// If client is nil, http.DefaultClient is used.
func NewOpenSearch(endpoint string, client *http.Client) *OpenSearch {
	if client == nil {
		client = http.DefaultClient
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	return &OpenSearch{BaseURL: endpoint, Client: client}
}

// DoWithRetry runs req with retries on transient failures (no response: connection refused, timeout).
// Returns the first response received; does not retry on 5xx (caller can handle). Caller must close resp.Body.
// For requests with a body, set req.GetBody so the body can be re-sent on retry.
func (os *OpenSearch) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			req = req.Clone(ctx)
			if req.GetBody != nil {
				body, _ := req.GetBody()
				req.Body = body
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt*100) * time.Millisecond):
			}
		}
		resp, err := os.Client.Do(req)
		if err != nil {
			lastErr = err
			if isRetryableErr(err) {
				continue
			}
			return nil, err
		}
		return resp, nil
	}
	return nil, lastErr
}

func isRetryableErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}
