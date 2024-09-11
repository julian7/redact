package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/julian7/redact/repo"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

const AZURE_KX_FILENAME = "azure.json"

var ErrAlreadyWritten = errors.New("secret already written to key vault")

type AzureConfig struct {
	KeyVaultURL string `json:"keyvault_url"`
	SecretName  string `json:"secret_name"`
}

func (conf *AzureConfig) secretsClient() (*azsecrets.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("authenticating to Azure: %w", err)
	}
	client, err := azsecrets.NewClient(conf.KeyVaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("authenticating to Key Vault secrets: %w", err)
	}

	return client, nil
}

func Print(redactRepo *repo.Repo) {
	conf, err := LoadAzureRefFromKX(redactRepo)
	if err != nil {
		return
	}
	fmt.Printf("Azure key vault %s in secret %q\n", conf.KeyVaultURL, conf.SecretName)
}

func LoadKey(ctx context.Context, redactRepo *repo.Repo) error {
	conf, err := LoadAzureRefFromKX(redactRepo)
	if err != nil {
		return fmt.Errorf("loading key: %w", err)
	}

	client, err := conf.secretsClient()
	if err != nil {
		return fmt.Errorf("loading key: %w", err)
	}

	gotSecret, err := client.GetSecret(ctx, conf.SecretName, "", nil)
	if err != nil {
		return fmt.Errorf("retrieving key while loading: %w", err)
	}

	if err := redactRepo.SecretKey.Import(bytes.NewBuffer([]byte(*gotSecret.Value))); err != nil {
		return fmt.Errorf("importing key: %w", err)
	}
	if err := redactRepo.SecretKey.Save(); err != nil {
		return fmt.Errorf("saving key: %w")
	}

	return nil
}

func SaveKey(ctx context.Context, redactRepo *repo.Repo) error {
	conf, err := LoadAzureRefFromKX(redactRepo)
	if err != nil {
		return fmt.Errorf("saving key: %w", err)
	}

	client, err := conf.secretsClient()
	if err != nil {
		return fmt.Errorf("saving key: %w", err)
	}

	contentType := "text/plain"
	val := bytes.Buffer{}
	redactRepo.SecretKey.Export(&val)

	gotSecret, err := client.GetSecret(ctx, conf.SecretName, "", nil)
	if err == nil && *gotSecret.Value == val.String() {
		return ErrAlreadyWritten
	}
	params := azsecrets.SetSecretParameters{
		ContentType: &contentType,
		Value:       to.Ptr(val.String()),
	}
	_, err = client.SetSecret(ctx, conf.SecretName, params, nil)
	if err != nil {
		return fmt.Errorf("cannot write secret %s: %w", conf.SecretName, err)
	}

	return nil
}

func LoadAzureRefFromKX(redactRepo *repo.Repo) (*AzureConfig, error) {
	fn, err := redactRepo.GetExchangeFilename(AZURE_KX_FILENAME, nil)
	if err != nil {
		return nil, fmt.Errorf("building azure key exchange filename: %w", err)
	}

	confFile, err := os.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("reading azure key exchange file: %w", err)
	}
	conf := &AzureConfig{}

	if err = json.Unmarshal(confFile, conf); err != nil {
		return nil, fmt.Errorf("parsing azure key exchange file: %w", err)
	}

	return conf, nil
}

func SaveAzureRefToKX(redactRepo *repo.Repo, vaultname string, secretname string) error {
	fn, err := redactRepo.GetExchangeFilename(AZURE_KX_FILENAME, nil)
	if err != nil {
		return fmt.Errorf("building azure key exchange filename: %w", err)
	}

	configWriter, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening azure key exchange file: %w", err)
	}
	defer configWriter.Close()

	vaulturl := fmt.Sprintf("https://%s.vault.azure.net", vaultname)
	if strings.HasPrefix(vaultname, "https://") {
		vaultParsedUrl, err := url.Parse(vaultname)
		if err != nil {
			fmt.Errorf("parsing vaultname url: %w", err)
		}
		vaulturl = vaultParsedUrl.String()
	}

	conf := AzureConfig{
		KeyVaultURL: vaulturl,
		SecretName:  secretname,
	}

	enc := json.NewEncoder(configWriter)
	if err = enc.Encode(conf); err != nil {
		return fmt.Errorf("encoding azure key exchange file contents: %w", err)
	}

	return nil
}
