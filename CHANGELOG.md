# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

No changes so far.

## [v0.9.0] - March 28, 2024

Added:

* Status subcommand detects if `.gitattributes` files are encrypted.

Removed:

* go-git dependency. It doubled executable size, while it was unable to fulfil all git-cli requirements.

Changed:

* Freshened dependencies with vulnerabilities (Critical: CVE-2023-49569, High: GHSA-9763-4f94-gfch, CVE-2023-49568, Moderate: CVE-2023-48795)

Fixed:

* False read-only file system error on loading secret key. The bug has been introduced with go-billy, not re-enabling read/write functionality upon chown.

## [v0.8.0] - December 8, 2023

Added:

* ARM64 version added to MacOS universal binary builder

Changed:

* Switch from cobra + viper to urfave/cli. We lose configurability via configuration files, but this is a more slick, precisely configurable option.
* Switch from afero to billy. This shaves off about 1MB from binary size, and it paves the way to switch to go-git.
* Rebuilt on go 1.21 and updated dependencies because of several CVEs for go 1.19 stdlib.

Fixed:

* Clearer error message when repo is not redacted
* Omit unnecessary error message when a new encrypted file is introduced
* Bail out from fixing keydir permissions if filesystem has no Change capability (when FS object doesn't implement billy.Change interface)

## [v0.7.1] - October 3, 2022

Added:

* Universal binary builder for MacOS

Fixed:

* Fix `redact unlock` identify `--exported-key` properly. Help strings have also been updated to communicate that these keys are indeed files, and they accept '-' as inputs.

## [v0.7.0] - October 2, 2022

Added:

* New command: `redact key export` to export redact private key in PEM format.
* New option: `redact unlock --exported-key <filename>` unlocks repository with exported, PEM-encrypted key.
* Standard input is accepted for `redact unlock --key` and `redact unlock --exported-key` commands by providing "-" as file name.

## [v0.6.0] - January 1, 2022

Breaking changes:

* Unlocking with GPG key moved to `redact unlock gpg` command.
* Unlocking with secret key needs argument: `redact unlock --key <secret key>`.
* instead of touching files, redact runs `git add --renormalize <toplevel>` to re-run clean filter.
* environment variable mapping prefixed with "REDACT_"

Changed:

* Renaming "master key" with "secret key", "master" branch with "main."
* Enable ChaCha20-Poly1305 encryption for platforms with no AES-NI instruction set. Use with setting `--type` with `redact git clean`, setting `git.clean.type` in configuration, or setting `REDACT_GIT_CLEAN_TYPE` environment variable.
* Internal: separated repo/file management from of key management

Fixed:

* Don't throw error on `redact generate` if key exchange dir doesn't exist
* `redact git clean` never received `--epoch`
* Honor epoch number for clean filter: `git status` showed files encrypted with older keys as being different in cloned repos. Now `redact git clean` takes the original file name as a parameter, and it can read its epoch number from git blob database.

## [v0.5.0] - October 21, 2021

Changed:

* Licensing updated to be able to choose between Blue Oak Model License (original but not widely adopted) or MIT
* Strict permission mode enforcement can be disabled. In certain situations, where POSIX and Windows permission setters are not available (like in WSL2 mounting from Windows), permission enforcement is not possible.
* Replaced deprecated golang.org/x/crypto/openpgp with supported github.com/ProtonMail/go-crypto/openpgp. This allows support for newer key formats like ED25519.

Fixed:

* Debug messages infiltrated into encrypted files, which were not detected as
  encrypted files anymore.

## [v0.4.5] - November 23, 2020

Fixed:

* generate: threw empty error on success
* list: unlocked repo is optional

## [v0.4.4] - June 17, 2020

Fixed:

* unlock: convey gpg error messages to users
* key permission detection: do a new fs stat after permission fix attempt
* git: GIT_DIR detection with old (1.8) git versions. It is still common on RHEL7-like environments.

Added:

* unlock: hint manual decrypting of gitlab keys on failure

Changed:

* Removed dependency to github.com/sirupsen/logrus, there's no need for a heavyweight logger.

## [v0.4.3] - February 16, 2020

Fixed:

* windows showstopper: permissions checking always failed. Removed check on windows platform.

Added:

* File permissions check: self-heal permissions if possible before reporting errors (windows: just self-heal, there are no checks ATM).

Changed:

* use goreleaser for full build and release

## [v0.4.2] - December 30, 2019

Changed:

* more output goes to log
* dropped upx: it comes at a great cost when called repeatedly

Fixed:

* unlock: unnecessary EOF reporting while GPG command runs well
* unlock: null pointer dereference regression from v0.4.0

## [v0.4.1] - December 22, 2019

Changed:

* errors handled with less ceremony: they don't show up the error twice, and usage when execution fails.

Fixed:

* v0.4.0 showstopper regressions in unlock and status commands
* key exchange directory's `.gitattributes` file check has not been rewritten upon modification

## [v0.4.0] - December 16, 2019

Fixed:

* stability: removed the majority of global variables (except non-changing ones)
* stability: removed all implicit method invocations (eg. init functions)
* stability: fulfil lint requirements (except secret key's sha1 hash chunk representation, as it's not security-related)
* scope: reduced runtime scope (eg. logging, environment detection) to runtime

## [v0.3.0] - August 11, 2019

Added:

* [SHOULDERS](SHOULDERS.md) file added to express my gratitude to all of the projects / people who helped setting this project together, in any way.
* Inline documentation added
* More strict args management for subcommands
* Allow unlocking with a copy of plaintext secret key

Changed:

* Change README, as v0.2.0 is already out and under testing.
* magefile niceties: "build" target for native OS, ".exe" extension for windows target, more precise version info from git

Fixed:

* Unlock silently selected the first matching GnuPG secret key if there were two matching items found.
* Secret key double save - regression from de751821a14ca4ea128332217823fa69dc121695
* Allow gpg passphrase prompt using gpg's `--pinentry-mode loopback` option

## [v0.2.0]: July 19, 2019

Changed:

* Typo fix in README.md
* `redact status` honors --repo option (to be even more quiet than --quiet)
* Added a lot of error checks to surely catch all errors
* Replace shell-based compile script with [magefile](https://magefile.org/)
* More stdlib use
* Refactor most testing utility methods to [tester](https://github.com/julian7/tester) library
* Knowledge of AES and HMAC keys and sizes are moved to encryptor only. Keys should not known about required components.
* Atomic secret key replacement on save

## [v0.1.0]: July 3, 2019

Initial release

Added:

* "redact" CLI tool compiled to 64bit OSes (Linux, Windows, MacOS)
* AES256-GCM96 file encryption with nonce computed with HMAC SHA-256
* filter and diff filters for seamless encoding of files
* multiple versions of secret key handling
* secret key distribution using OpenPGP keys

[Unreleased]: https://github.com/julian7/redact
[v0.9.0]: https://github.com/julian7/redact/releases/tag/v0.9.0
[v0.8.0]: https://github.com/julian7/redact/releases/tag/v0.8.0
[v0.7.1]: https://github.com/julian7/redact/releases/tag/v0.7.1
[v0.7.0]: https://github.com/julian7/redact/releases/tag/v0.7.0
[v0.6.0]: https://github.com/julian7/redact/releases/tag/v0.6.0
[v0.5.0]: https://github.com/julian7/redact/releases/tag/v0.5.0
[v0.4.5]: https://github.com/julian7/redact/releases/tag/v0.4.5
[v0.4.4]: https://github.com/julian7/redact/releases/tag/v0.4.4
[v0.4.3]: https://github.com/julian7/redact/releases/tag/v0.4.3
[v0.4.2]: https://github.com/julian7/redact/releases/tag/v0.4.2
[v0.4.1]: https://github.com/julian7/redact/releases/tag/v0.4.1
[v0.4.0]: https://github.com/julian7/redact/releases/tag/v0.4.0
[v0.3.0]: https://github.com/julian7/redact/releases/tag/v0.3.0
[v0.2.0]: https://github.com/julian7/redact/releases/tag/v0.2.0
[v0.1.0]: https://github.com/julian7/redact/releases/tag/v0.1.0
