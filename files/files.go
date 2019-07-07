package files

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/julian7/redact/encoder"
	"github.com/pkg/errors"
)

const (
	// FileMagic magic string the encoded file starts with
	FileMagic = "\000REDACTED\000"
)

type fileHeader struct {
	Preamble [10]byte
	Encoding uint32
	Epoch    uint32
}

// Encode encodes an IO stream into another IO stream
func (k *MasterKey) Encode(encodingFormat uint32, epoch uint32, reader io.Reader, writer io.Writer) error {
	key, err := k.Key(epoch)
	if err != nil {
		return errors.Wrap(err, "encoding stream")
	}
	enc, err := encoder.NewEncoder(int(encodingFormat), key.AES(), key.HMAC())
	if err != nil {
		return errors.Wrap(err, "setting up encoder")
	}
	in, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "reading input stream")
	}
	out, err := enc.Encode(in)
	if err != nil {
		return errors.Wrap(err, "encoding stream")
	}
	header := fileHeader{Encoding: encodingFormat, Epoch: epoch}
	copy(header.Preamble[:], FileMagic)
	err = binary.Write(writer, binary.BigEndian, header)
	if err != nil {
		return errors.Wrap(err, "writing file header")
	}
	_, err = writer.Write(out)
	return errors.Wrap(err, "writing encoded stream")
}

// Decode encodes an IO stream into another IO stream
func (k *MasterKey) Decode(reader io.Reader, writer io.Writer) error {
	var header fileHeader
	err := k.readHeader(reader, &header)
	if err != nil {
		return err
	}
	key, err := k.Key(header.Epoch)
	if err != nil {
		return errors.Wrap(err, "retrieving key")
	}
	enc, err := encoder.NewEncoder(int(header.Encoding), key.AES(), key.HMAC())
	if err != nil {
		return errors.Wrap(err, "setting up encoder")
	}
	in, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "reading stream")
	}
	out, err := enc.Decode(in)
	if err != nil {
		return errors.Wrap(err, "decoding stream")
	}
	_, err = writer.Write(out)
	return errors.Wrap(err, "writing decoded stream")
}

// FileStatus returns file encryption status and key used
func (k *MasterKey) FileStatus(reader io.Reader) (bool, uint32) {
	var header fileHeader
	err := k.readHeader(reader, &header)
	if err != nil {
		return false, 0
	}
	return true, header.Epoch
}

func (k *MasterKey) readHeader(reader io.Reader, header *fileHeader) error {
	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		return errors.Wrap(err, "reading file header")
	}
	if !bytes.Equal(header.Preamble[:], []byte(FileMagic)) {
		return errors.New("invalid file preamble")
	}
	return nil
}
