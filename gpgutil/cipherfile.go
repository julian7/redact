package gpgutil

import (
	"crypto"
	"fmt"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// Encrypt encrypts stream into an OpenPGP-encrypted stream
func Encrypt(reader io.Reader, writer io.Writer, receiver *openpgp.Entity) error {
	hints := openpgp.FileHints{IsBinary: true}
	config := packet.Config{
		DefaultHash:            crypto.SHA256,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionZLIB,
		CompressionConfig: &packet.CompressionConfig{
			Level: 9,
		},
		RSABits: 4096,
	}

	plain, err := openpgp.Encrypt(writer, []*openpgp.Entity{receiver}, nil, &hints, &config)
	if err != nil {
		return fmt.Errorf("creating encryption stream: %w", err)
	}

	defer plain.Close()

	if _, err = io.Copy(plain, reader); err != nil {
		return fmt.Errorf("writing secret key to encryption stream: %w", err)
	}

	return nil
}
