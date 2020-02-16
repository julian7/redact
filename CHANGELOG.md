# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

No changes so far.

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
* stability: fulfil lint requirements (except master key's sha1 hash chunk representation, as it's not security-related)
* scope: reduced runtime scope (eg. logging, environment detection) to runtime

## [v0.3.0] - August 11, 2019

Added:

* [SHOULDERS](SHOULDERS.md) file added to express my gratitude to all of the projects / people who helped setting this project together, in any way.
* Inline documentation added
* More strict args management for subcommands
* Allow unlocking with a copy of plaintext master file

Changed:

* Change README, as v0.2.0 is already out and under testing.
* magefile niceties: "build" target for native OS, ".exe" extension for windows target, more precise version info from git

Fixed:

* Unlock silently selected the first matching GnuPG secret key if there were two matching items found.
* Master key double save - regression from de751821a14ca4ea128332217823fa69dc121695
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
* Atomic master key replacement on save

## [v0.1.0]: July 3, 2019

Initial release

Added:

* "redact" CLI tool compiled to 64bit OSes (Linux, Windows, MacOS)
* AES256-GCM96 file encryption with nonce computed with HMAC SHA-256
* filter and diff filters for seamless encoding of files
* multiple versions of secret key handling
* secret key distribution using OpenPGP keys

[Unreleased]: https://github.com/julian7/redact
[v0.4.3]: https://github.com/julian7/redact/releases/tag/v0.4.3
[v0.4.2]: https://github.com/julian7/redact/releases/tag/v0.4.2
[v0.4.1]: https://github.com/julian7/redact/releases/tag/v0.4.1
[v0.4.0]: https://github.com/julian7/redact/releases/tag/v0.4.0
[v0.3.0]: https://github.com/julian7/redact/releases/tag/v0.3.0
[v0.2.0]: https://github.com/julian7/redact/releases/tag/v0.2.0
[v0.1.0]: https://github.com/julian7/redact/releases/tag/v0.1.0
