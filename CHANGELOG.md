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

## [v1.6.0] - 2019-03-29

### Added

- Enable trading on centralized exchanges ([505162a86777f99fba26bc953b3125aba90e2f7e](https://github.com/stellar/kelp/commit/505162a86777f99fba26bc953b3125aba90e2f7e)) ([a9ab0346ddd3500395018d1dbcf426200b5fb112](https://github.com/stellar/kelp/commit/a9ab0346ddd3500395018d1dbcf426200b5fb112)) ([3a1b4c467495a5ebb8219c554dd8a5e4d63723e5](https://github.com/stellar/kelp/commit/3a1b4c467495a5ebb8219c554dd8a5e4d63723e5))
- support for OrderConstraintsFilter to preempt invalid orders ([b9ba73071d97e9e0e8c6f61989ff9375be4dbbeb](https://github.com/stellar/kelp/commit/b9ba73071d97e9e0e8c6f61989ff9375be4dbbeb))
- Expand CCXT exchanges enabled on Kelp, including trading-enabled exchanges ([0631bb1ec8892e331907614cca94d66aed3ee026](https://github.com/stellar/kelp/commit/0631bb1ec8892e331907614cca94d66aed3ee026)) ([40c56416e07a5f974815ee0ca11992c8825e57c2](https://github.com/stellar/kelp/commit/40c56416e07a5f974815ee0ca11992c8825e57c2))
- Fill Tracker should accommodate N errors before causing bot to exit ([de12bfe16e7d68f76798e4b99f60cb005386c2cb](https://github.com/stellar/kelp/commit/de12bfe16e7d68f76798e4b99f60cb005386c2cb))

### Changed

- Use FeeStats() method from new horizonclient package in stellar/go repo; run `glide install` to update vendored dependencies, _do NOT run `glide up` since that will break the dependencies installed because of an issue with how glide works_ ([ce226cc20ce6a38fe56728c91432db9edd7cb272](https://github.com/stellar/kelp/commit/ce226cc20ce6a38fe56728c91432db9edd7cb272))
- krakenExchange logs better error message when trading pair is missing ([807139ff2b4b6fb81726459b1ef1958d95f7cd95](https://github.com/stellar/kelp/commit/807139ff2b4b6fb81726459b1ef1958d95f7cd95))
- Sort exchanges based on what is tested first ([8d6d4502032b8b5b237ad9d2a6fe3c38f176c541](https://github.com/stellar/kelp/commit/8d6d4502032b8b5b237ad9d2a6fe3c38f176c541))

### Fixed

- Fixed Mirror Strategy not working without offsetTrades flag ([09f76e891967a146363ccbd8fe8ccf53656c270e](https://github.com/stellar/kelp/commit/09f76e891967a146363ccbd8fe8ccf53656c270e))
- Fix minVolume amounts when offsetting trades in mirror strategy ([0576aa1724b98b3d359b1d82ca36b156b0251977](https://github.com/stellar/kelp/commit/0576aa1724b98b3d359b1d82ca36b156b0251977))

## [v1.5.0] - 2019-03-04

### Added

- support for dynamic fee calculations using the `/fee_stats` endpoint ([c0f7b5de726b7718f9335ba6b1e1aceca3d9a72d](https://github.com/stellar/kelp/commit/c0f7b5de726b7718f9335ba6b1e1aceca3d9a72d))
- include ccxt-kraken as a read-only exchange ([796ae5964a835ca441bb67f5964656dc2b5ecdb4](https://github.com/stellar/kelp/commit/796ae5964a835ca441bb67f5964656dc2b5ecdb4))
- USDC as a recognized asset ([3894b9a38fef601e5fe27a901f3f66a2071478f2](https://github.com/stellar/kelp/commit/3894b9a38fef601e5fe27a901f3f66a2071478f2))

### Changed

- load orderConstraints for ccxtExchange from CCXT API ([8d28c11b488e74f04d23f2ef62df6b603e731c68](https://github.com/stellar/kelp/commit/8d28c11b488e74f04d23f2ef62df6b603e731c68))

## [v1.4.0] - 2019-02-06

### Added
- Support to run Kelp in maker-only mode using the trader.cfg file ([081aa210e684678b94c0ec2d772ad808eec9f0d6](https://github.com/stellar/kelp/commit/081aa210e684678b94c0ec2d772ad808eec9f0d6))
- Support for an SDEX priceFeed so you can follow ticker prices from other SDEX markets ([8afec86c831c45aef2e4cc8e0c85c1de6d192325](https://github.com/stellar/kelp/commit/8afec86c831c45aef2e4cc8e0c85c1de6d192325))

## [v1.3.0] - 2019-01-10

### Added
- mirror strategy offsets trades onto the backing exchange, run `glide up` to update dependencies ([3a703a359db541b636cab38c3dd8a7fbe6df7193](https://github.com/stellar/kelp/commit/3a703a359db541b636cab38c3dd8a7fbe6df7193))
- ccxt integration now supports trading APIs for all exchanges ([5cf0aedc67eff89a8f82082326f878844ac7b5d5](https://github.com/stellar/kelp/commit/5cf0aedc67eff89a8f82082326f878844ac7b5d5))
- randomized delay via the MAX_TICK_DELAY_MILLIS ([4b74affb9933bf08a093ee66cea46c1b3fb87753](https://github.com/stellar/kelp/commit/4b74affb9933bf08a093ee66cea46c1b3fb87753))

### Changed
- balanced strategy avoids unncessary re-randomization on every update cycle ([0be414c77c2f12c9b4b624922aea5841e84c704c](https://github.com/stellar/kelp/commit/0be414c77c2f12c9b4b624922aea5841e84c704c))

### Fixed
- fix op_underfunded issue when hitting capacity limits ([d339e421f82de9e2996e45e71d745d81dff2f3f0](https://github.com/stellar/kelp/commit/d339e421f82de9e2996e45e71d745d81dff2f3f0))

## [v1.2.0] - 2018-11-26

### Added
- support for alerting with PagerDuty as the first implementation, run `glide up` to update the dependency ([5e46ae0d94751d85dbb2e8f73094f5d96af0df5e](https://github.com/stellar/kelp/commit/5e46ae0d94751d85dbb2e8f73094f5d96af0df5e))
- support for logging to a file with the `--log` or `-l` command-line option followed by the prefix of the log filename
- support for basic monitoring with a health check service, run `glide up` to update the dependency ([c6374c35cff9dfa46da342aa5342f312dcd337c4](https://github.com/stellar/kelp/commit/c6374c35cff9dfa46da342aa5342f312dcd337c4))
- `iter` command line param to run for only a fixed number of iterations, run `glide up` to update the dependencies
- new DELETE_CYCLES_THRESHOLD config value in trader config file to allow some tolerance of errors before deleting all offers ([f2537cafee8d620e1c4aabdd3d072d90628801b8](https://github.com/stellar/kelp/commit/f2537cafee8d620e1c4aabdd3d072d90628801b8))

### Changed
- reduced the number of available assets that are recognized by the GetOpenOrders() API for Kraken
- levels are now logged with prices in the quote asset and amounts in the base asset for the sell, buysell, and balanced strategies
- clock tick is now synchronized at the start of each cycle ([cd33d91b2d468bfbce6d38a6186d12c86777b7d5](https://github.com/stellar/kelp/commit/cd33d91b2d468bfbce6d38a6186d12c86777b7d5))

### Fixed
- conversion of asset symbols in the GetOpenOrders() API for Kraken, reducing the number of tested asset symbols with this API
- fix op_underfunded errors when we hit capacity limits for non-XLM assets ([e6bebee9aeadf6e00a829a28c125f5dffad8c05c](https://github.com/stellar/kelp/commit/e6bebee9aeadf6e00a829a28c125f5dffad8c05c))

## [v1.1.2] - 2018-10-30

### Added
- log balance with liabilities

### Changed
- scripts/build.sh: update VERSION format and LDFLAGS to include git branch info

### Fixed
- fix op_underfunded errors when we hit capacity limits

## [v1.1.1] - 2018-10-22

### Fixed
- fixed bot panicing when it cannot cast ticker bid/ask values to a float64 from CCXT's FetchTicker endpoint (0ccbc495e18b1e3b207dad5d3421c7556c63c004) ([issue #31](https://github.com/stellar/kelp/issues/31))

## [v1.1.0] - 2018-10-19

### Added
- support for [CCXT](https://github.com/ccxt/ccxt) via [CCXT-REST API](https://github.com/franz-see/ccxt-rest), increasing exchange integrations for priceFeeds and mirroring [diff](https://github.com/stellar/kelp/compare/0db8f2d42580aa87867470e428d5f0f63eed5ec6^...33bc7b98418129011b151d0f56c9c0770a3d897e)

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
- support for [CAP-0003](https://github.com/stellar/stellar-protocol/blob/master/core/cap-0003.md) introduced in stellar-core protocol v10 ([issue #2](https://github.com/stellar/kelp/issues/2))

## v1.0.0-rc1 - 2018-08-13

### Added
- Kelp bot with a few basic strategies, priceFeeds, and support for integrating with the Kraken Exchange.
- Modular design allowing anyone to plug in their own strategies
- Robust logging
- Configuration file based approach to setting up a bot
- Documentation on existing capabilities

[Unreleased]: https://github.com/stellar/kelp/compare/v1.6.0...HEAD
[v1.6.0]: https://github.com/stellar/kelp/compare/v1.5.0...v1.6.0
[v1.5.0]: https://github.com/stellar/kelp/compare/v1.4.0...v1.5.0
[v1.4.0]: https://github.com/stellar/kelp/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/stellar/kelp/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/stellar/kelp/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/stellar/kelp/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/stellar/kelp/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/stellar/kelp/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/stellar/kelp/compare/v1.0.0-rc3...v1.0.0
[v1.0.0-rc3]: https://github.com/stellar/kelp/compare/v1.0.0-rc2...v1.0.0-rc3
[v1.0.0-rc2]: https://github.com/stellar/kelp/compare/v1.0.0-rc1...v1.0.0-rc2
