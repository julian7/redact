package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type Config struct {
	KeyvaultURL string
	SecretName  string
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
		case "vault", "keyvault":
			var vaulturl string
			if strings.HasPrefix(val, "https://") {
				vaulturl = val
			} else {
				vaulturl = fmt.Sprintf("https://%s.vault.azure.net", val)
			}

			parsedURL, err := url.Parse(vaulturl)
			if err != nil {
				return nil, fmt.Errorf("parsing vault url: %w", err)
			}

			config.KeyvaultURL = parsedURL.String()
		case "secret":
			config.SecretName = val
		default:
			fmt.Printf("unknown setting: %v\n", key)
		}
	}

	if config.KeyvaultURL == "" {
		return nil, ErrMissingKeyvault
	}

	if config.SecretName == "" {
		return nil, ErrMissingSecret
	}

	return config, nil
}

func (conf *Config) secretsClient() (*azsecrets.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("authenticating to Azure: %w", err)
	}

	client, err := azsecrets.NewClient(conf.KeyvaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("authenticating to Key Vault secrets: %w", err)
	}

	return client, nil
}
