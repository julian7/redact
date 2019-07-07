package encoder_test

import (
	"bytes"
	"testing"

	"github.com/julian7/redact/encoder"

	"github.com/pkg/errors"
)

type fakeEncoder struct {
	aes  []byte
	hmac []byte
}

func newFakeEncoder(aes, hmac []byte) (encoder.Encoder, error) {
	return &fakeEncoder{aes: aes, hmac: hmac}, nil
}
func (e *fakeEncoder) Encode([]byte) ([]byte, error) { return nil, errors.New("not implemented") }
func (e *fakeEncoder) Decode([]byte) ([]byte, error) { return nil, errors.New("not implemented") }

func TestRegisterEncoder(t *testing.T) {
	id := 10
	aes := []byte("foo")
	hmac := []byte("bar")
	err := encoder.RegisterEncoder(id, newFakeEncoder)
	if err != nil {
		t.Error(err)
		return
	}
	enc, err := encoder.NewEncoder(id, aes, hmac)
	if err != nil {
		t.Error(err)
		return
	}
	fakeEnc, ok := enc.(*fakeEncoder)
	if !ok {
		t.Errorf("unexpected received encoder: %T", enc)
	}
	if !bytes.Equal(fakeEnc.aes, aes) {
		t.Errorf("AES key doesn't match.\nExpected: %q\nReceived: %q", aes, fakeEnc.aes)
	}
	if !bytes.Equal(fakeEnc.hmac, hmac) {
		t.Errorf("HMAC key doesn't match.\nExpected: %q\nReceived: %q", hmac, fakeEnc.hmac)
	}
	if err := encoder.UnregisterEncoder(id); err != nil {
		t.Errorf("unregistering encoder: %v", err)
	}
}

func TestNewEncoderError(t *testing.T) {
	id := 15
	aes := []byte("foo")
	hmac := []byte("bar")
	_, err := encoder.NewEncoder(id, aes, hmac)
	if err == nil || err.Error() != "invalid encoding type 15" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegisterEncoderError(t *testing.T) {
	id := 20
	err := encoder.RegisterEncoder(id, newFakeEncoder)
	if err != nil {
		t.Error(err)
		return
	}
	err = encoder.RegisterEncoder(id, newFakeEncoder)
	if err == nil || err.Error() != "encoder type 20 already exists" {
		t.Errorf("unexpected error: %v", err)
		return
	}
}

func TestUnegisterEncoderError(t *testing.T) {
	id := 25
	err := encoder.UnregisterEncoder(id)
	if err == nil || err.Error() != "encoder type 25 doesn't exist" {
		t.Errorf("unexpected error: %v", err)
		return
	}
}
