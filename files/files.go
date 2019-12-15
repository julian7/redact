package files

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/julian7/redact/encoder"
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
		return fmt.Errorf("encoding stream: %w", err)
	}

	enc, err := encoder.NewEncoder(int(encodingFormat), key.Secret())
	if err != nil {
		return fmt.Errorf("setting up encoder: %w", err)
	}

	in, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("reading input stream: %w", err)
	}

	out, err := enc.Encode(in)
	if err != nil {
		return fmt.Errorf("encoding stream: %w", err)
	}

	header := fileHeader{Encoding: encodingFormat, Epoch: epoch}
	copy(header.Preamble[:], FileMagic)

	err = binary.Write(writer, binary.BigEndian, header)
	if err != nil {
		return fmt.Errorf("writing file header: %w", err)
	}

	_, err = writer.Write(out)
	if err != nil {
		return fmt.Errorf("writing encoded stream: %w", err)
	}

	return nil
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
		return fmt.Errorf("retrieving key: %w", err)
	}

	enc, err := encoder.NewEncoder(int(header.Encoding), key.Secret())
	if err != nil {
		return fmt.Errorf("setting up encoder: %w", err)
	}

	in, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("reading stream: %w", err)
	}

	out, err := enc.Decode(in)
	if err != nil {
		return fmt.Errorf("decoding stream: %w", err)
	}

	_, err = writer.Write(out)
	if err != nil {
		return fmt.Errorf("writing decoded stream: %w", err)
	}

	return nil
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
		return fmt.Errorf("reading file header: %w", err)
	}

	if !bytes.Equal(header.Preamble[:], []byte(FileMagic)) {
		return errors.New("invalid file preamble")
	}

	return nil
}
