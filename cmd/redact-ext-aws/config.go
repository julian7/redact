package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type Config struct {
	KeyID     string
	ParamPath string
}

func loadConfig(args []string) (*Config, error) {
	config := &Config{}

	for idx, item := range args {
		before, after, ok := strings.Cut(item, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: %w", idx+1, ErrInvalidArgument)
		}

		key := before
		val := after

		switch key {
		case "keyid":
			config.KeyID = val
		case "param":
			config.ParamPath = val
		default:
			fmt.Printf("unknown setting: %v\n", key)
		}
	}

	if config.KeyID == "" {
		return nil, ErrMissingKeyID
	}

	if config.ParamPath == "" {
		return nil, ErrMissingParamPath
	}

	return config, nil
}

func ssmClient(ctx context.Context) (*ssm.Client, error) {
	options := [](func(*config.LoadOptions) error){}

	awscfg, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	return ssm.NewFromConfig(awscfg), err
}

func (config *Config) get(ctx context.Context, client *ssm.Client) (*ssm.GetParameterOutput, error) {
	return client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           new(config.ParamPath),
		WithDecryption: new(true),
	})
}

func (config *Config) put(ctx context.Context, client *ssm.Client, key string) (*ssm.PutParameterOutput, error) {
	return client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      new(config.ParamPath),
		Value:     new(key),
		DataType:  new("text"),
		Overwrite: new(true),
		KeyId:     new(config.KeyID),
		Type:      types.ParameterTypeSecureString,
	})
}
