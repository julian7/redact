package gpgutil

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

var errFinalWrite = errors.New("final write failed")

type finalWriter struct {
	bytes.Buffer
	fail bool
}

func (w *finalWriter) Write(p []byte) (int, error) {
	if w.fail || bytes.Contains(p, []byte("-----END")) {
		return 0, errFinalWrite
	}

	return w.Buffer.Write(p)
}

func (w *finalWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

type eofReader struct {
	io.Reader
	writer *finalWriter
}

func (r *eofReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if errors.Is(err, io.EOF) {
		r.writer.fail = true
	}

	return n, err
}

func TestWriteFinalizationErrors(t *testing.T) {
	key, err := openpgp.NewEntity("recipient", "", "recipient@example.com", &packet.Config{RSABits: 1024})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("encryption", func(t *testing.T) {
		writer := &finalWriter{}

		reader := &eofReader{Reader: strings.NewReader("secret"), writer: writer}
		if err := Encrypt(reader, writer, key); !errors.Is(err, errFinalWrite) {
			t.Fatalf("expected finalization error, got %v", err)
		}
	})
	t.Run("armor", func(t *testing.T) {
		if err := SavePubKey(&finalWriter{}, key, true); !errors.Is(err, errFinalWrite) {
			t.Fatalf("expected armor footer error, got %v", err)
		}
	})
}
