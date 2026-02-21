package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Addr        string `envconfig:"ADDR" default:":8080"`
	AWSRegion   string `envconfig:"AWS_REGION" default:"us-east-1"`
	DynamoTable string `envconfig:"DYNAMO_TABLE" required:"true"`
	JWTSecret   string `envconfig:"JWT_SECRET" required:"true"`
}

func Load() (*Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return nil, err
	}
	return &c, nil
}
