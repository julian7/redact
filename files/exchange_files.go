package files

import (
	"fmt"
	"path/filepath"
)

const (
	// ExtKeyArmor is public key ASCII armor file extension in Key Exchange folder
	ExtKeyArmor = ".asc"
	// ExtSecret is encrypted secret key file extension in Key Exchange folder
	ExtSecret = ".key"
)

// GetExchangeFilenameStubFor returns file name stub of the Key Exchange for an
// OpenPGP key identified by its full public key ID.
//
// Add extensions for files:
//
// - .asc: Public key ASCII armor file
// - .key: Secret key encryped with public key
func (k *SecretKey) GetExchangeFilenameStubFor(fingerprint []byte) (string, error) {
	kxdir := ExchangeDir(k.RepoInfo.Toplevel)
	if err := k.ensureExchangeDir(kxdir); err != nil {
		return "", err
	}

	return filepath.Join(kxdir, fmt.Sprintf("%x", fingerprint)), nil
}

// ExchangePubKeyFile returns full filename for Public key ASCII armor
func ExchangePubKeyFile(stub string) string {
	return fmt.Sprintf("%s%s", stub, ExtKeyArmor)
}

// ExchangeSecretKeyFile returns full filename for Secret key exchange
func ExchangeSecretKeyFile(stub string) string {
	return fmt.Sprintf("%s%s", stub, ExtSecret)
}
