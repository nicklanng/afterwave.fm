package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sopatech/afterwave.fm/internal/infra"
)

const (
	bcryptCost = 12
	jwtExpiry  = 7 * 24 * time.Hour
)

var (
	ErrEmailTaken   = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
)

type Service interface {
	Signup(ctx context.Context, email, password string) (userID string, err error)
	Login(ctx context.Context, email, password string) (userID string, err error)
	DeleteAccount(ctx context.Context, userID string) error
	NewToken(userID string) (string, error)
}

type service struct {
	db        *infra.Dynamo
	table     string
	jwtSecret []byte
}

func NewService(db *infra.Dynamo, table string, jwtSecret []byte) Service {
	return &service{db: db, table: table, jwtSecret: jwtSecret}
}

type userRow struct {
	ID           string `dynamodbav:"id"`
	Email        string `dynamodbav:"email"`
	PasswordHash string `dynamodbav:"password_hash"`
	CreatedAt    string `dynamodbav:"created_at"`
}

func (s *service) Signup(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	if email == "" || len(password) < 8 {
		return "", fmt.Errorf("email and password (min 8 chars) required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	// Check email not already registered (GSI email-index)
	check, err := s.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.table),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return "", err
	}
	if len(check.Items) > 0 {
		return "", ErrEmailTaken
	}

	userID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	row := userRow{
		ID:           userID,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
	}

	item, err := attributevalue.MarshalMap(row)
	if err != nil {
		return "", err
	}

	_, err = s.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.table),
		Item:      item,
	})
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	if email == "" {
		return "", ErrInvalidCreds
	}

	out, err := s.db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.table),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil || len(out.Items) == 0 {
		return "", ErrInvalidCreds
	}

	var row userRow
	if err := attributevalue.UnmarshalMap(out.Items[0], &row); err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCreds
	}
	return row.ID, nil
}

func (s *service) DeleteAccount(ctx context.Context, userID string) error {
	_, err := s.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.table),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: userID},
		},
	})
	return err
}

func (s *service) NewToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(jwtExpiry)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.jwtSecret)
}

func (s *service) ValidateToken(tokenString string) (userID string, err error) {
	tok, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(*jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := tok.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return "", errors.New("invalid token")
	}
	return claims.Subject, nil
}

func normalizeEmail(s string) string {
	// minimal: trim and lowercase
	b := []byte(s)
	start := 0
	for start < len(b) && (b[start] == ' ' || b[start] == '\t') {
		start++
	}
	end := len(b)
	for end > start && (b[end-1] == ' ' || b[end-1] == '\t') {
		end--
	}
	s = string(b[start:end])
	// lowercase
	var out []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out = append(out, c)
	}
	return string(out)
}
