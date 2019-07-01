# To be implemented

- key exchange
  - use GnuPG for decryption
  - use [golang.org/x/crypto/openpgp](https://godoc.org/golang.org/x/crypto/openpgp) for encryption
  - store public GPG keys into key exchange files
  - allow keybase's encryption to be used
- fix encryption issues
  - encrypted what shouldn't been
  - not encrypted which should have been
  - re-encrypt with a newer key
