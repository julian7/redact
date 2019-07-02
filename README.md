# Redact, data encryptor in git

**PLEASE NOTE**: this project is still under heavy development, and not suitable for normal use. Also, most of the features described here are still don't exist.

Redact allows you to store sensitive data in a git repository by encrypting / decrypting it on-the-fly.

The project is very similar to [git-crypt](https://github.com/AGWA/git-crypt), [transcrypt](https://github.com/elasticdog/transcrypt), and [git-secret](https://github.com/sobolevn/git-secret).

## WARNING

In nominal cases, you should never store your secrets in a git repository. Use a different tool for that. However, when the sensitivity levels of secret / plaintext files are not too far away (like general and specific settings for a closed source application), it's a nice touch to have an extra layer of security on top of what a git service can provide.

## What Redact provides, what other tools don't?

* Redact is written in go, and as such, it can bring encryption into environments with no bash (???), and can be cross-compiled.
* it uses OpenPGP for key exchange (almost like git-crypt)
* it requires GnuPG only for unlocking, all OpenPGP operations are handled internally
* it stores OpenPGP keys of collaborators next to their encrypted master key copies
* it supports key rotation (like transcrypt)
* auto-generates key (not like transcrypt)
* only dependencies are git and GnuPG
* it doesn't commit into the repository automatically

## Security

Keys are prepared for AES-256 (256 bits) and HMAC-256 (512 bits). The system is able to store multiple keys, with different epoch numbers. All keys are usable for decryption, but by default only the latest key is used for encryption.

File encoding uses AES256-GCM96 encoding. Encryption nonce is calculated from the plaintext's HMAC key, taking the first 96 bits. Encrypted file stores this calculated nonce, and ciphertext. During decoding, the saved nonce is checked against calculated HMAC, whether its first 96 bits are matching.

## Subcommands

* [x] key: master key commands:
  * [x] init: initializes master key
  * [x] generate: generates new master key
  * [x] info (default): shows master key info
  * [x] list: lists all keys
* [ ] lock: locks repository (deletes local key and removes diff/filter configs)
* [x] unlock: unlocks repository with local key
* [o] access: key exchange commands
  * [x] ls/list: list user access
  * [x] grant: add GPG key access
  * [ ] revoke: remove GPG key access
* [x] git: git filter commands
  * [x] clean: acts as clean filter for git
  * [x] diff: acts as diff filter for git
  * [x] smudge: acts as smudge filter for git
* [x] status: list files' encryption status

See [the to do](TODO.md) file for details.

## Revoke access

When you lose trust of someone, there is one thing we can't do: we can't revoke
historical access to secrets stored in the repository. We can do something about the future though:

* remove encrypted key from the key exchange folder. This will stop the user from recreating the master key in the future.
* generate a new key: rotate master key. You can re-encrypt secret files as they are, but since usually the untrusted party already knows about the secrets, they can easily figure out the newly encrypted files are indeed having the same content. This can possibly help them learning about the new secret key.
* replacing secrets: when encrypted files are supposed to be exposed, the best thing we can do is not just replacing their encryptions, but replacing secrets too. For example, if encrypted files are secret parts of key pairs (like a TLS certificate), we might want to revoke the full certificate altogether, generating new ones.

As always, play safe, and revoke all secrets if there is any chance it can cause damage.
