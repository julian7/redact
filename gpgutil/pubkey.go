package gpgutil

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// PrintKey prints information about an OpenPGP key entity
func PrintKey(key *openpgp.Entity) {
	fmt.Printf(
		"KeyID: %s, fingerprint: %x\n",
		key.PrimaryKey.KeyIdShortString(),
		key.PrimaryKey.Fingerprint,
	)

	for _, id := range key.Identities {
		var expires string

		if id.SelfSignature.SigType != packet.SigTypePositiveCert {
			continue
		}

		sig := id.SelfSignature

		if sig.KeyLifetimeSecs == nil {
			expires = "no expiration"
		} else {
			expiry := sig.CreationTime.Add(time.Duration(*sig.KeyLifetimeSecs) * time.Second)
			if !expiry.After(time.Now()) {
				expires = fmt.Sprintf("%s (expired)", expiry)
			} else {
				expires = expiry.String()
			}
		}

		fmt.Printf(
			"  identity: %s, expires: %s\n",
			id.Name,
			expires,
		)
	}
}

// LoadPubKeyFromFile Loads public key into openpgp's Entity
func LoadPubKeyFromFile(path string, armor bool) (openpgp.EntityList, error) {
	out, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening pgp asc file %s: %w", path, err)
	}

	defer out.Close()

	entries, err := LoadPubKey(out, armor)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	return entries, nil
}

// LoadPubKey loads public key from a readable stream
func LoadPubKey(reader io.Reader, armor bool) (openpgp.EntityList, error) {
	var entities openpgp.EntityList

	var err error

	if armor {
		entities, err = openpgp.ReadArmoredKeyRing(reader)
	} else {
		entities, err = openpgp.ReadKeyRing(reader)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"reading keyring of pgp%s: %w",
			map[bool]string{false: "", true: " armor"}[armor],
			err,
		)
	}

	return entities, nil
}

// SavePubKey saves a public keys from an entity to a stream
func SavePubKey(raw io.Writer, key *openpgp.Entity, isArmor bool) error {
	var writer io.Writer

	if isArmor {
		arm, err := armor.Encode(raw, openpgp.PublicKeyType, make(map[string]string))
		if err != nil {
			return fmt.Errorf("creating armor stream: %w", err)
		}
		defer arm.Close()
		writer = arm
	} else {
		writer = raw
	}

	if err := key.Serialize(writer); err != nil {
		return fmt.Errorf("serializing public key to exchange store: %w", err)
	}

	return nil
}
