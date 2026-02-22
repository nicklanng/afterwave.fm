package infra

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/guregu/dynamo/v2"
)

// Dynamo embeds dynamo.DB for table access. Use Table(name) for Get/Put/Delete/WriteTx.
// Client() returns the underlying DynamoDB API for DescribeTable/CreateTable (e.g. in tests).
type Dynamo struct {
	*dynamo.DB
}

// NewDynamo creates a DynamoDB client using guregu/dynamo. If endpointURL is non-empty
// (e.g. http://localhost:8001 for DynamoDB Local), the client uses that endpoint.
func NewDynamo(ctx context.Context, region, endpointURL string) (*Dynamo, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	opts := []func(*dynamodb.Options){}
	if endpointURL != "" {
		u, err := url.Parse(endpointURL)
		if err != nil {
			return nil, err
		}
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(u.String())
		})
	}
	return &Dynamo{DB: dynamo.New(cfg, opts...)}, nil
}
