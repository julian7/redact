package files

import (
	"fmt"
	"path/filepath"
)

const (
	// ExtKeyArmor is public key ASCII armor file extension in Key Exchange folder
	ExtKeyArmor = ".asc"
	// ExtMaster is encrypted master key file extension in Key Exchange folder
	ExtMaster = ".key"
)

// GetExchangeFilenameStubFor returns file name stub of the Key Exchange for an
// OpenPGP key identified by its full public key ID.
//
// Add extensions for files:
//
// - .asc: Public key ASCII armor file
// - .key: Master key encryped with public key
func (k *MasterKey) GetExchangeFilenameStubFor(toplevel string, fingerprint [20]byte) (string, error) {
	kxdir, err := k.getExchangeDir(toplevel)
	if err != nil {
		return "", err
	}
	return filepath.Join(kxdir, fmt.Sprintf("%x", fingerprint)), nil
}

// ExchangePubKeyFile returns full filename for Public key ASCII armor
func ExchangePubKeyFile(stub string) string {
	return fmt.Sprintf("%s%s", stub, ExtKeyArmor)
}

// ExchangeMasterKeyFile returns full filename for Master key exchange
func ExchangeMasterKeyFile(stub string) string {
	return fmt.Sprintf("%s%s", stub, ExtMaster)
}
