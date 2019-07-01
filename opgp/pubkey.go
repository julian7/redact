package opgp

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

type PubKeyInfo struct {
	PubKey *packet.PublicKey
	UserID *packet.UserId
}

// ReadPubkeys reads OpenPGP public keys from ASCII armor, and returns public key packets
func ReadPubkeys(key []byte) ([]*PubKeyInfo, error) {
	var keys []*PubKeyInfo
	var lastKey int
	in := bytes.NewReader(key)
	blk, err := armor.Decode(in)
	if err != nil {
		return nil, errors.Wrap(err, "decoding armor")
	}
	if blk.Type != openpgp.PublicKeyType {
		return nil, errors.New("not a public key")
	}
	reader := packet.NewReader(blk.Body)
	for {
		pkt, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrap(err, "reading public key")
		}
		pubKey, ok := pkt.(*packet.PublicKey)
		if ok {
			keys = append(keys, &PubKeyInfo{PubKey: pubKey})
			lastKey++
			continue
		}
		userID, ok := pkt.(*packet.UserId)
		if ok {
			keys[lastKey-1].UserID = userID
			continue
		}
		_, ok = pkt.(*packet.Signature)
		if ok {
			// we are not interested in signatures
			continue
		}
		return nil, errors.Errorf("item is not a public key: %+v", pkt)
	}
	return keys, nil
}
