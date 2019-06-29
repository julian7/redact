# Redact, data encryptor in git

**PLEASE NOTE**: this project is still under heavy development, and not suitable for normal use. Also, most of the features described here are still don't exist.

Redact allows you to store sensitive data in a git repository by encrypting / decrypting it on-the-fly.

The project is very similar to [git-crypt](https://github.com/AGWA/git-crypt), [transcrypt](https://github.com/elasticdog/transcrypt), and [git-secret](https://github.com/sobolevn/git-secret).

## WARNING

In nominal cases, you should never store your secrets in a git repository. Use a different tool for that. However, when secrets are integral part of the repo (like if it is all about storing secret information), it's a nice touch to have an extra layer of security on top of what a git service can provide.

## What Redact provides, what other tools don't?

* Redact is written in go, and as such, it can bring encryption into environments with no bash (???), and can be cross-compiled.
* it uses GPG for key exchange (à la git-crypt)
* it supports key rotation (like transcrypt)
* auto-generates key (not like transcrypt)
* only dependencies are git and gnupg, and this latter is only for key exchange

## Security

Keys are prepared for AES-256 (256 bits) and HMAC-256 (512 bits). The system is able to store multiple keys, with different epoch numbers. All keys are usable for decryption, but by default only the latest key is used for encryption.

File encoding uses AES256-GCM96 encoding, which uses an HMAC-based Extract-and-Expand Key Derivation Function for calculating the final AES / HMAC keys. Encryption nonce
is calculated from the plaintext's HMAC key, taking the first 96 bits. Encrypted file stores this calculated nonce, and ciphertext. During decoding, the saved nonce is checked against calculated HMAC, whether its first 96 bits are matching.

## Subcommands

* key: master key commands:
  * init: initializes master key
  * generate: generates new master key
  * info (default): shows master key info
  * list: lists all keys
* lock: deletes local key
* unlock: unlocks repository with local key
* access: key exchange commands
  * ls/list: list user access
  * grant: add GPG key access
  * revoke: remove GPG key access
* git: git filter commands
  * clean: acts as clean filter for git
  * diff: acts as diff filter for git
  * smudge: acts as smudge filter for git
* ls/list: list files' encryption status

## Revoke access

When you lose trust of someone, there is one thing we can't do: we can't revoke
historical access to secrets stored in the repository. We can do something about the future though:

* remove encrypted key from the key exchange folder. This will stop the user from recreating the master key in the future.
* rekey: rotate master key. It implies re-encrypting all the secret files with the new key. This will enforce everyone to recreate the master key from key exchange, but the one we've removed won't be able to unlock the new key.
* replacing secrets: when encrypted files are supposed to be exposed, the best thing we can do is not just replacing their encryptions, but replacing secrets too. For example, if encrypted files are secret parts of key pairs (like a TLS certificate), we might want to revoke the full certificate altogether, generating new ones.

As always, play safe, and revoke all secrets if there is any chance it can cause damage.
