# Redact, data encryptor in git

**PLEASE NOTE**: this project is not released yet, and it's under testing. Use it for your own risk.

Redact allows you to store sensitive data in a git repository by encrypting / decrypting it on-the-fly.

The project is very similar to [git-crypt](https://github.com/AGWA/git-crypt), [transcrypt](https://github.com/elasticdog/transcrypt), and [git-secret](https://github.com/sobolevn/git-secret).

## WARNING

In nominal cases, you should never store your secrets in a git repository. Use a different tool for that. However, when the sensitivity levels of secret / plaintext files are not too far away (like general and specific settings for a closed source application), it's a nice touch to have an extra layer of security on top of what a git service can provide.

## Intro

In order to store secrets in a git repository, you have to initialize it:

```shell
$ redact init
New repo key created: .git/redact (#1 c90014bb)
```

This creates a secret key into `.git/redact` directory, and it also sets up diff / filter attributes for later use.

Then, tell the repo which files are going to be encrypted. Create a `.gitattributes` file like this one:

```text
*.key filter=redact diff=redact
```

This file will instruct git to encrypt every file with the `.key` extension. Let's create a secret file called `private.key`:

```text
Secret Information
```

Then, add them to version control:

```shell
$ git add --all
$ git status -sb
## No commits yet on main
A  .gitattributes
A  private.key
$ git commit -m initial
[main (root-commit) 0b9eb27] initial
 2 files changed, 2 insertions(+)
 create mode 100644 .gitattributes
 create mode 100644 private.key
```

At this point we have an encrypted file in history:

```shell
$ redact status
            .gitattributes
encrypted:  private.key
There are 0 files to be fixed.
There are 0 files to be re-encrypted.
```

Just check whether it has been encrypted:

```text
$ git ls-tree HEAD
100644 blob 8fa0eac4c7076b9f10c99460c2feaf905f0e5cd2  .gitattributes
100644 blob a24b3ae46b388e8ca0011a13d79f94cdd65feb35  private.key
$ git cat-file blob a24b3ae46b388e8ca0011a13d79f94cdd65feb35 | hexdump -C
00000000  00 52 45 44 41 43 54 45  44 00 00 00 00 00 00 00  |.REDACTED.......|
00000010  00 01 3f ef ae 3f 90 aa  73 cd 39 8c 89 1c 2e e8  |..?..?..s.9.....|
00000020  dc 20 22 66 ea 7a 97 7f  b2 86 38 7e 40 fc 05 e5  |. "f.z....8~@...|
00000030  5a 44 41 c5 37 8e c4 8c  02 28 a4 ca 38 6a 76 7b  |ZDA.7....(..8jv{|
00000040  08                                                |.|
00000041
$ git cat-file blob a24b3ae46b388e8ca0011a13d79f94cdd65feb35 | redact git smudge
Secret Information
$ cat private.key
Secret Information
```

Add contributors:

```text
$ redact access grant keybase.io/julian7
KeyID: BDE0F1CE, fingerprint: 1857918cd0b4d303071d6624466cbb98bde0f1ce
  identity: keybase.io/julian7 <js@iksz.hu>, expires: 2027-10-08 18:45:56 +0200 CEST
  identity: keybase.io/julian7 <julian7@keybase.io>, expires: 2025-01-07 22:35:45 +0100 CET
INFO[0000] Added 1 key. Don't forget to commit exchange files to the repository.
$ git add --all
$ git status -sb
## main
A  .redact/.gitattributes
A  .redact/1857918cd0b4d303071d6624466cbb98bde0f1ce.asc
A  .redact/1857918cd0b4d303071d6624466cbb98bde0f1ce.key
```

## What Redact provides, what other tools don't?

* Redact is written in go, and as such, it can bring encryption into environments with no bash (???), and can be cross-compiled.
* it uses OpenPGP for key exchange (almost like git-crypt)
* it requires GnuPG only for unlocking, all OpenPGP operations are handled internally
* it stores OpenPGP keys of collaborators next to their encrypted secret key copies
* it supports key rotation (like transcrypt)
* auto-generates key (not like transcrypt)
* only dependencies are git and GnuPG
* it doesn't commit into the repository automatically

## Security

Keys are prepared for AES-256 (256 bits) and HMAC-256 (512 bits). The system is able to store multiple keys, with different epoch numbers. All keys are usable for decryption, but by default only the latest key is used for encryption.

File encoding uses AES256-GCM96 encoding. Encryption nonce is calculated from the plaintext's HMAC key, taking the first 96 bits. Encrypted file stores this calculated nonce, and ciphertext. During decoding, the saved nonce is checked against calculated HMAC, whether its first 96 bits are matching.

Alternatively, ChaCha20-Poly1305 can be used in place of AES256-GCM96. According to the [Automatic cipher suite ordering in crypto/tls](https://go.dev/blog/tls-cipher-suites) blog entry in the Go Blog, AES128-GCM96 is not suitable for the post-quantumcomputing world, which is still just fine for TLS encryption, but not in our case. However, AES256-GCM96 provides adequate protection for that case too. On the other hand, the consensus on using AES+GCM with no hardware acceleration is pretty much discouraged as it's very hard to implement it effectively and securely.

Therefore, if encryption is used in environments which lacks hardware AES acceleration, it's more secure to use ChaCha20-Poly1305 instead.

To switch, either set `REDACT_GIT_CLEAN_TYPE` environment variable, or set `git.clean.type` in configuration to `chacha20-poly1305` (case insensitive).

## Subcommands

* key: secret key commands:
  * init: initializes secret key
  * export: exports secret key in a PEM-encoded (readable) format
  * generate: generates new secret key
  * info (default): shows secret key info
  * list: lists all keys
* lock: locks repository (deletes local key and removes diff/filter configs)
* unlock: unlocks repository with local key
  * gpg: unlocks repository with GPG-encrypted key from key exchange
* access: key exchange commands
  * ls/list: list user access
  * grant: add OpenPGP key access
  * update: re-encrypt secret key with OpenPGP keys (not implemented yet)
* git: git filter commands
  * clean: acts as clean filter for git
  * diff: acts as diff filter for git
  * smudge: acts as smudge filter for git
* status: list files' encryption status

See [the to do](TODO.md) file for details.

## Revoke access

When you lose trust of someone, there is one thing we can't do: we can't revoke
historical access to secrets stored in the repository. We can do something about the future though:

* remove encrypted key from the key exchange folder. This will stop the user from recreating the secret key in the future.
* generate a new key: rotate secret key. You can re-encrypt secret files as they are, but since usually the untrusted party already knows about the secrets, they can easily figure out the newly encrypted files are indeed having the same content. This can possibly help them learning about the new secret key.
* replacing secrets: when encrypted files are supposed to be exposed, the best thing we can do is not just replacing their encryptions, but replacing secrets too. For example, if encrypted files are secret parts of key pairs (like a TLS certificate), we might want to revoke the full certificate altogether, generating new ones.

As always, play safe, and revoke all secrets if there is any chance it can cause damage.

## Configuration with environment variables

A couple of options are configurable with environment variables too, which are taking precedence over command-line options. You can set the following variables:

* `REDACT_GIT_CLEAN_EPOCH`: sets `--epoch` option for `redact git clean` subcommand
* `REDACT_GIT_CLEAN_TYPE`: sets `--type` option for `redact git clean` subcommand
* `REDACT_LOG_LEVEL`: sets `--verbosity` global option
* `REDACT_LOG_LEVEL`: sets `--logfile` global option
* `REDACT_STRICT`: sets `--strict-permissions` global option
* `REDACT_UNLOCK_EXPORTED_KEY`: sets `--exported-key` option for `redact unlock` subcommand
* `REDACT_UNLOCK_GPG_KEY`: sets `--gpgkey` option for `redact unlock gpg` subcommand
* `REDACT_UNLOCK_KEY`: sets `--key` option for `redact unlock` subcommand

## Any issues?

Open a ticket, perhaps a pull request. We support [GitHub Flow](https://guides.github.com/introduction/flow/). You might want to [fork](https://guides.github.com/activities/forking/) this project first.

## Licensing

SPDX-License-Identifier: BlueOak-1.0.0 OR MIT

This software is licensed under two licenses of your choice: Blue Oak Public License 1.0, or MIT Public License.
