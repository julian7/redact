package keys

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
	err := binary.Read(reader, binary.BigEndian, &header)
	if err != nil {
		return errors.Wrap(err, "reading file header")
	}
	if bytes.Compare(header.Preamble[:], []byte(FileMagic)) != 0 {
		return errors.New("invalid file preamble")
	}
	key, err := k.Key(header.Epoch)
	if err != nil {
		return errors.Wrap(err, "retrieving key")
	}
	enc, err := encoder.NewEncoder(int(header.Encoding), key.AES(), key.HMAC())
	if err != nil {
		return errors.Wrap(err, "retrieving encoder")
	}
	in, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}
	out, err := enc.Decode(in)
	if err != nil {
		return errors.Wrap(err, "decoding stream")
	}
	_, err = writer.Write(out)
	return errors.Wrap(err, "writing decoded stream")
}
