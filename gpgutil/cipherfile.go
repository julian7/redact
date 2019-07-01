package gpgutil

import (
	"crypto"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
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
		return errors.Wrap(err, "creating encryption stream")
	}
	defer plain.Close()
	_, err = io.Copy(plain, reader)
	if err != nil {
		return errors.Wrap(err, "writing master key to encryption stream")
	}
	return nil
}
