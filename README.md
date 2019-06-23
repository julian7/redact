# Redact, data encryptor in git

Redact allows you to store sensitive data in a git repository by encrypting / decrypting it on-the-fly.

The project is very similar to [git-crypt](https://github.com/AGWA/git-crypt), [transcrypt](https://github.com/elasticdog/transcrypt), and [git-secret](https://github.com/sobolevn/git-secret).

## What Redact provides, what other tools don't?

* Redact is written in go, and as such, it can bring encryption into environments with no bash (???), and can be cross-compiled.
* it uses GPG for key exchange (à la git-crypt)
* it supports key rotation (like transcrypt)
* auto-generates key (not like transcrypt)
* only dependencies are git and gnupg, and this latter is only for key exchange

## Subcommands

* key: master key commands:
  * init: initializes master key
  * rekey: reinitializes master key, and re-encrypts files
  * info (default): shows master key info
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
