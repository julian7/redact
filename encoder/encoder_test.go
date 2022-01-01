package encoder_test

import (
	"crypto/cipher"
	"errors"
	"testing"

	"github.com/julian7/redact/encoder"
	"github.com/julian7/tester"
)

type fakeEncoder struct {
	key []byte
}

func newFakeEncoder(key []byte) (encoder.AEAD, error) {
	return &fakeEncoder{key: key}, nil
}

func (e *fakeEncoder) AEAD() (cipher.AEAD, error) { return nil, errors.New("not implemented") }
func (e *fakeEncoder) KeySize() int               { return 0 }
func (e *fakeEncoder) String() string             { return "fake-encoder" }

func TestRegisterEncoder(t *testing.T) {
	id := 10

	tt := []struct {
		name string
		key  string
		err  error
	}{
		{name: "empty key", key: "", err: encoder.ErrKeyTooSmall},
		{name: "short key", key: "foo", err: encoder.ErrKeyTooSmall},
		{name: "appropriate keysize", key: "0123456789012345678901234567890123456789012345678901234567890123", err: nil},
		{name: "long key", key: "0123456789012345678901234567890123456789012345678901234567890123xtraXTRA", err: nil},
	}

	for _, tc := range tt {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := encoder.RegisterEncoder(id, newFakeEncoder)
			if err != nil {
				t.Error(err)

				return
			}

			_, err = encoder.NewEncoder(id, []byte(tc.key))
			if err := tester.AssertError(tc.err, err); err != nil {
				t.Error(err)
			}

			if err := encoder.UnregisterEncoder(id); err != nil {
				t.Errorf("unregistering encoder: %v", err)
			}
		})
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
