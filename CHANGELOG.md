# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [v1.1.1] - 2018-10-22

### Fixed
- fixed bot panicing when it cannot cast ticker bid/ask values to a float64 from CCXT's FetchTicker endpoint (0ccbc495e18b1e3b207dad5d3421c7556c63c004) ([issue #31](https://github.com/lightyeario/kelp/issues/31))

## [v1.1.0] - 2018-10-19

### Added
- support for [CCXT](https://github.com/ccxt/ccxt) via [CCXT-REST API](https://github.com/franz-see/ccxt-rest), increasing exchange integrations for priceFeeds and mirroring [diff](https://github.com/lightyeario/kelp/compare/0db8f2d42580aa87867470e428d5f0f63eed5ec6^...33bc7b98418129011b151d0f56c9c0770a3d897e)

## [v1.0.0] - 2018-10-15

### Changed
- executables for windows should use the .exe extension (7b5bbc9eb5b776a27c63483c4af09ca38937670d)

### Fixed
- fixed divide by zero error (fa7d7c4d5a2a256d6cfcfe43a65e530e3c06862e)

## [v1.0.0-rc3] - 2018-09-29

### Added
- support for all currencies available on Kraken

## [v1.0.0-rc2] - 2018-09-28

### Added
- This CHANGELOG file

### Changed
- Updated dependency `github.com/stellar/go` to latest version `5bbd27814a3ffca9aeffcbd75a09a6164959776a`, run `glide up` to update this dependency

### Fixed
- If `SOURCE_SECRET_SEED` is missing or empty then the bot will not crash now.
- support for [CAP-0003](https://github.com/stellar/stellar-protocol/blob/master/core/cap-0003.md) introduced in stellar-core protocol v10 ([issue #2](https://github.com/lightyeario/kelp/issues/2))


## v1.0.0-rc1 - 2018-08-13

### Added
- Kelp bot with a few basic strategies, priceFeeds, and support for integrating with the Kraken Exchange.
- Modular design allowing anyone to plug in their own strategies
- Robust logging
- Configuration file based approach to setting up a bot
- Documentation on existing capabilities

[Unreleased]: https://github.com/lightyeario/kelp/compare/v1.1.1...HEAD
[v1.1.1]: https://github.com/lightyeario/kelp/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/lightyeario/kelp/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/lightyeario/kelp/compare/v1.0.0-rc3...v1.0.0
[v1.0.0-rc3]: https://github.com/lightyeario/kelp/compare/v1.0.0-rc2...v1.0.0-rc3
[v1.0.0-rc2]: https://github.com/lightyeario/kelp/compare/v1.0.0-rc1...v1.0.0-rc2
