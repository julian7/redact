package encoder_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/julian7/redact/encoder"
)

var (
	//go:embed fixtures/key.txt
	sampleKey []byte
	//go:embed fixtures/nonce.txt
	sampleNonce []byte
	//go:embed fixtures/plaintext.txt
	samplePlaintext []byte
	//go:embed fixtures/ciphertext-aes256-gcm96.bin
	sampleAESCiphertext []byte
	//go:embed fixtures/ciphertext-chacha20-poly1305.bin
	sampleChaChaCiphertext []byte
)

func TestEncode(t *testing.T) {
	tt := []struct {
		name    string
		factory func([]byte) (encoder.AEAD, error)
		cipher  []byte
	}{
		{name: "AES256-GCM96", factory: encoder.NewAES256GCM96, cipher: sampleAESCiphertext},
		{name: "ChaCha20-Poly1305", factory: encoder.NewChaCha20Poly1305, cipher: sampleChaChaCiphertext},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			enc, err := tc.factory(sampleKey)
			if err != nil {
				t.Errorf("cannot create encoder: %v", err)

				return
			}

			aead, err := enc.AEAD()
			if err != nil {
				t.Errorf("cannot get AEAD: %v", err)

				return
			}

			ret := aead.Seal(nil, sampleNonce[:aead.NonceSize()], samplePlaintext, nil)
			if !bytes.Equal(ret, tc.cipher) {
				// o, err := os.Create("fixtures/" + tc.name + "-sample.bin")
				// if err != nil {
				// 	t.Fatal(err)
				// }
				// _, err = o.Write(ret)
				// if err != nil {
				// 	t.Fatal(err)
				// }
				// o.Close()
				t.Errorf("Encrypted message not matching: %x", ret)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tt := []struct {
		name    string
		factory func([]byte) (encoder.AEAD, error)
		cipher  []byte
	}{
		{name: "AES256-GCM96", factory: encoder.NewAES256GCM96, cipher: sampleAESCiphertext},
		{name: "ChaCha20-Poly1305", factory: encoder.NewChaCha20Poly1305, cipher: sampleChaChaCiphertext},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			enc, err := tc.factory(sampleKey)
			if err != nil {
				t.Errorf("cannot create encoder: %v", err)

				return
			}

			aead, err := enc.AEAD()
			if err != nil {
				t.Errorf("cannot get AEAD: %v", err)

				return
			}

			ret, err := aead.Open(nil, sampleNonce[:aead.NonceSize()], tc.cipher, nil)
			if err != nil {
				t.Errorf("cannot decode: %v", err)

				return
			}

			if !bytes.Equal(ret, samplePlaintext) {
				t.Errorf("Encrypted message not matching: %q", ret)
			}
		})
	}
}
