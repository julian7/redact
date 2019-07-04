# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

No changes so far.

## [v0.1.0]: July 3, 2019

Initial release

### Added

* "redact" CLI tool compiled to 64bit OSes (Linux, Windows, MacOS)
* AES256-GCM96 file encryption with nonce computed with HMAC SHA-256
* filter and diff filters for seamless encoding of files
* multiple versions of secret key handling
* secret key distribution using OpenPGP keys

[Unreleased]: https://github.com/julian7/redact
[v0.1.0]: https://github.com/julian7/redact/releases/tag/v0.1.0
