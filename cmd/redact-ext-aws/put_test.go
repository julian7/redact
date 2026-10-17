package main

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/julian7/redact/ext"
)

type fakeParameterStore struct {
	value  string
	writes int
	fail   bool
}

func (store *fakeParameterStore) GetParameter(
	context.Context, *ssm.GetParameterInput, ...func(*ssm.Options),
) (*ssm.GetParameterOutput, error) {
	return &ssm.GetParameterOutput{Parameter: &types.Parameter{Value: &store.value}}, nil
}

func (store *fakeParameterStore) PutParameter(
	_ context.Context, input *ssm.PutParameterInput, _ ...func(*ssm.Options),
) (*ssm.PutParameterOutput, error) {
	if !*input.Overwrite || input.Type != types.ParameterTypeSecureString || *input.KeyId != "kms-key" {
		return nil, errors.New("storage settings changed")
	}

	if store.fail {
		return nil, errors.New("write failed")
	}

	store.value = *input.Value
	store.writes++

	return &ssm.PutParameterOutput{}, nil
}

func TestPublicationConverges(t *testing.T) {
	config := &Config{KeyID: "kms-key", ParamPath: "/redact"}

	store := &fakeParameterStore{}
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
			if err := cmd.Run(context.Background(), []string{"redact-ext-aws", tc.operation}); ext.ExitCode(err) != tc.code {
				t.Fatalf("%s: got %v, want %d", tc.operation, err, tc.code)
			}
		})
	}
}
