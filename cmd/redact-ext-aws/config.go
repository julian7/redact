package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go/ptr"
)

type Config struct {
	KeyID     string
	ParamPath string
}

func loadConfig(args []string) (*Config, error) {
	config := &Config{}

	for idx, item := range args {
		i := strings.Index(item, "=")
		if i < 0 {
			return nil, fmt.Errorf("line %d: %w", idx+1, ErrInvalidArgument)
		}

		key := item[:i]
		val := item[i+1:]

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
		Name:           ptr.String(config.ParamPath),
		WithDecryption: ptr.Bool(true),
	})
}

func (config *Config) put(ctx context.Context, client *ssm.Client, key string) (*ssm.PutParameterOutput, error) {
	return client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      ptr.String(config.ParamPath),
		Value:     ptr.String(key),
		DataType:  ptr.String("text"),
		Overwrite: ptr.Bool(true),
		KeyId:     ptr.String(config.KeyID),
		Type:      types.ParameterTypeSecureString,
	})
}
