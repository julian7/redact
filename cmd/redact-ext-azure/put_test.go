package main

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/julian7/redact/ext"
)

type fakeSecretStore struct {
	value  string
	writes int
	fail   bool
}

func (store *fakeSecretStore) GetSecret(
	context.Context, string, string, *azsecrets.GetSecretOptions,
) (azsecrets.GetSecretResponse, error) {
	return azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: &store.value}}, nil
}

func (store *fakeSecretStore) SetSecret(
	_ context.Context, _ string, input azsecrets.SetSecretParameters, _ *azsecrets.SetSecretOptions,
) (azsecrets.SetSecretResponse, error) {
	if *input.ContentType != "text/plain" {
		return azsecrets.SetSecretResponse{}, errors.New("storage settings changed")
	}

	if store.fail {
		return azsecrets.SetSecretResponse{}, errors.New("write failed")
	}

	store.value = *input.Value
	store.writes++

	return azsecrets.SetSecretResponse{}, nil
}

func TestPublicationConverges(t *testing.T) {
	config := &Config{KeyvaultURL: "https://fixture.vault.azure.net", SecretName: "redact"}

	store := &fakeSecretStore{}
	for range 2 {
		if err := config.save(context.Background(), store, "exported PEM"); err != nil {
			t.Fatal(err)
		}
	}

	if store.value != "exported PEM" || store.writes != 1 {
		t.Fatalf("retry changed stored key: %+v", store)
	}

	store.fail = true
	if err := config.save(context.Background(), store, "rotated PEM"); ext.ExitCode(err) != ext.Failure {
		t.Fatalf("write failure hidden: %v", err)
	}
}

func TestCommandOutcomes(t *testing.T) {
	for _, tc := range []struct {
		operation string
		code      int
	}{
		{"grant", ext.NotApplicable}, {"revoke", ext.NotApplicable}, {"unknown", ext.NotApplicable},
		{"get", ext.MissingData}, {"put", ext.MissingData}, {"update", ext.MissingData},
	} {
		t.Run(tc.operation, func(t *testing.T) {
			cmd := app()

			cmd.Writer, cmd.ErrWriter = io.Discard, io.Discard
			if err := cmd.Run(context.Background(), []string{"redact-ext-azure", tc.operation}); ext.ExitCode(err) != tc.code {
				t.Fatalf("%s: got %v, want %d", tc.operation, err, tc.code)
			}
		})
	}
}
