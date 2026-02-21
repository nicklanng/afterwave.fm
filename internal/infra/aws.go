package infra

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Dynamo struct {
	Client *dynamodb.Client
}

func NewDynamo(ctx context.Context, region string) (*Dynamo, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &Dynamo{Client: dynamodb.NewFromConfig(cfg)}, nil
}
