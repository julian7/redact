package encoder_test

import (
	"bytes"
	"testing"

	"github.com/julian7/redact/encoder"

	"github.com/pkg/errors"
)

type fakeEncoder struct {
	key []byte
}

func newFakeEncoder(key []byte) (encoder.Encoder, error) {
	return &fakeEncoder{key: key}, nil
}
func (e *fakeEncoder) Encode([]byte) ([]byte, error) { return nil, errors.New("not implemented") }
func (e *fakeEncoder) Decode([]byte) ([]byte, error) { return nil, errors.New("not implemented") }

func TestRegisterEncoder(t *testing.T) {
	id := 10
	key := []byte("foo")

	err := encoder.RegisterEncoder(id, newFakeEncoder)
	if err != nil {
		t.Error(err)
		return
	}

	enc, err := encoder.NewEncoder(id, key)
	if err != nil {
		t.Error(err)
		return
	}

	fakeEnc, ok := enc.(*fakeEncoder)
	if !ok {
		t.Errorf("unexpected received encoder: %T", enc)
	}

	if !bytes.Equal(fakeEnc.key, key) {
		t.Errorf("Secret key doesn't match.\nExpected: %q\nReceived: %q", key, fakeEnc.key)
	}

	if err := encoder.UnregisterEncoder(id); err != nil {
		t.Errorf("unregistering encoder: %v", err)
	}
}

func TestNewEncoderError(t *testing.T) {
	id := 15
	key := []byte("foo")

	_, err := encoder.NewEncoder(id, key)
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
