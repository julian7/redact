package keyv0_test

import (
	"testing"

	keyv0 "github.com/julian7/redact/files/key_v0"
)

const (
	sampleAES  = "0123456789abcdefghijklmnopqrstuv"
	sampleHMAC = "0123456789abcdefghijklmnopqrstuv0123456789abcdefghijklmnopqrstuv"
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
	if len(key.AES()) != keyv0.AESKeySize {
		t.Errorf(
			"Invalid AES key size\nExpected: %d\nReceived: %d",
			keyv0.AESKeySize,
			len(key.AES()),
		)
	}
	if len(key.HMAC()) != keyv0.HMACKeySize {
		t.Errorf(
			"Invalid HMAC key size\nExpected: %d\nReceived: %d",
			keyv0.HMACKeySize,
			len(key.HMAC()),
		)
	}
}

func TestString(t *testing.T) {
	key := keyv0.KeyV0{
		Epoch: 1,
	}
	copy(key.AESData[:], sampleAES)
	copy(key.HMACData[:], sampleHMAC)
	strval := key.String()
	expected := "#1 b3ec9ddd"
	if strval != expected {
		t.Errorf(
			"Invalid string representation of key\nExpected: %s\nReceived: %s",
			expected,
			strval,
		)
	}
}
