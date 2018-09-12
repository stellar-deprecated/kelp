# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- This CHANGELOG file

### Changed
- Updated dependency `github.com/stellar/go` to latest version `5bbd27814a3ffca9aeffcbd75a09a6164959776a`, run `glide up` to update this dependency

### Deprecated

### Removed

### Fixed
- If `SOURCE_SECRET_SEED` is missing or empty then the bot will not crash now.

### Security


## v1.0.0-rc1 - 2018-08-13

### Added
- Kelp bot with a few basic strategies, priceFeeds, and support for integrating with the Kraken Exchange.
- Modular design allowing anyone to plug in their own strategies
- Robust logging
- Configuration file based approach to setting up a bot
- Documentation on existing capabilities

[Unreleased]: https://github.com/lightyeario/kelp/compare/v1.0.0-rc1...HEAD
