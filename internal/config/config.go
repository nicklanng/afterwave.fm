package config

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr        string `envconfig:"ADDR" default:":8080"`
	AWSRegion   string `envconfig:"AWS_REGION" default:"us-east-1"`
	DynamoTable string `envconfig:"DYNAMO_TABLE" required:"true"`
	JWTSecret   string `envconfig:"JWT_SECRET" required:"true" obfuscate:"true"`
}

func Load() (*Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// ObfuscateStr returns a masked version of s for safe logging (e.g. secrets).
func ObfuscateStr(s string) string {
	if len(s) <= 2 {
		return "**"
	}
	if len(s) < 8 {
		return s[:1] + "****" + s[len(s)-1:]
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// LogConfigVars logs each field of config that has an envconfig tag.
// Fields with struct tag obfuscate:"true" are masked via ObfuscateStr.
// config must be a struct or pointer to struct.
func LogConfigVars(logger *slog.Logger, config any) {
	v := reflect.ValueOf(config)
	t := reflect.TypeOf(config)

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		logger.Error("config must be a struct")
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		envTag := field.Tag.Get("envconfig")
		if envTag == "" {
			continue
		}
		envKey := strings.Fields(envTag)[0]

		shouldObfuscate := field.Tag.Get("obfuscate") == "true"

		var valueStr string
		switch fieldValue.Kind() {
		case reflect.String:
			valueStr = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valueStr = fmt.Sprintf("%d", fieldValue.Int())
		case reflect.Bool:
			valueStr = fmt.Sprintf("%t", fieldValue.Bool())
		case reflect.Float32, reflect.Float64:
			valueStr = fmt.Sprintf("%.2f", fieldValue.Float())
		default:
			valueStr = fmt.Sprintf("%v", fieldValue.Interface())
		}

		if shouldObfuscate {
			valueStr = ObfuscateStr(valueStr)
		}

		logger.Info("config", "var", envKey, "value", valueStr)
	}
}
