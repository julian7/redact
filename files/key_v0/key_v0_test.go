package keyv0_test

import (
	"testing"

	keyv0 "github.com/julian7/redact/files/key_v0"
)

const (
	sampleCode = "0123456789abcdefghijklmnopqrstuv"
)

func TestNewKey(t *testing.T) {
	epoch := uint32(5)
	key := keyv0.NewKey(epoch)
	if key == nil {
		t.Error("No key returned")
		return
	}
	if key.Epoch != epoch {
		t.Errorf(
			"invalid epoch\nExpected: %d\nReceived: %d",
			epoch,
			key.Epoch,
		)
	}
}

func TestKeyVersion(t *testing.T) {
	key := keyv0.KeyV0{}
	expected := uint32(0)
	if key.Version() != expected {
		t.Errorf(
			"invalid version\nExpected: %d\nReceived: %d",
			expected,
			key.Version(),
		)
	}
}

func TestKeyType(t *testing.T) {
	key := keyv0.KeyV0{}
	if key.Type() != 0 {
		t.Errorf("invalid key format type: %d", key.Type())
	}
}

func TestGenerate(t *testing.T) {
	key := keyv0.KeyV0{Epoch: 1}
	err := key.Generate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(key.Secret()) != keyv0.SecretSize {
		t.Errorf(
			"Invalid AES key size\nExpected: %d\nReceived: %d",
			keyv0.SecretSize,
			len(key.Secret()),
		)
	}
}

func TestString(t *testing.T) {
	key := keyv0.KeyV0{
		Epoch: 1,
	}
	copy(key.SecretData[:], sampleCode+sampleCode+sampleCode)
	strval := key.String()
	expected := "#1 0006b8c0"
	if strval != expected {
		t.Errorf(
			"Invalid string representation of key\nExpected: %s\nReceived: %s",
			expected,
			strval,
		)
	}
}
