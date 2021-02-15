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



## [v1.11.0] - 2020-02-15

### Added

- mirror max base volume cap ([#556](https://github.com/stellar/kelp/issues/556))
- log time taken for update loop ([#558](https://github.com/stellar/kelp/issues/558))
- add pprof experimental cli option ([12ac3ce9d4d27acd57da0f9d6edeecdf671e1f4f](https://github.com/stellar/kelp/commit/12ac3ce9d4d27acd57da0f9d6edeecdf671e1f4f))
- Enable GUI metrics tracking (part of [#508](https://github.com/stellar/kelp/issues/508), [07e8b1e294f026ec7e12964775fcd2b1a3a56df8](https://github.com/stellar/kelp/commit/07e8b1e294f026ec7e12964775fcd2b1a3a56df8))
- Add buy infrastructure to volume filter (part of [#522](https://github.com/stellar/kelp/issues/522))
- Bitstamp Integration ([#489](https://github.com/stellar/kelp/issues/489))
- Add metrics for operation counts (part of [#551](https://github.com/stellar/kelp/issues/551))
- Add Pull Request Guidelines ([#601](https://github.com/stellar/kelp/issues/601))
- sleep mode type ([#606](https://github.com/stellar/kelp/issues/606))
- significant reliability improvement in Kelp GUI with regards to errors from backend to frontend ([002a726c877555b277076e280cb32f32ba650af0](https://github.com/stellar/kelp/commit/002a726c877555b277076e280cb32f32ba650af0))
- add utils.MustParseAsset helper function ([e65e14006d9c32e7349d4d7e23ffe68cede0a8e5](https://github.com/stellar/kelp/commit/e65e14006d9c32e7349d4d7e23ffe68cede0a8e5))
- new buyTwap strategy ([#522](https://github.com/stellar/kelp/issues/522))
- Implement missing filter logic related to buy side ([#636](https://github.com/stellar/kelp/issues/636))
- Kelp GUI: enable public network ([#649](https://github.com/stellar/kelp/issues/649))

### Changed

- network speedup: check markets cache for existing symbols in ccxt.go#symbolExists() ([#559](https://github.com/stellar/kelp/issues/559))
- improve condition for placeSellOpsFirst in mirror strategy ([94a30d652f31d125f8b8424472e8c42e321fbe94](https://github.com/stellar/kelp/commit/94a30d652f31d125f8b8424472e8c42e321fbe94))
- update circleci config to replace quote asset for test runs ([7a15ab6e1656d51cd7bdf7bc5c9654c439024bfe](https://github.com/stellar/kelp/commit/7a15ab6e1656d51cd7bdf7bc5c9654c439024bfe))
- conditionally reset cached balances and liabilities to reduce network calls, closes [#561](https://github.com/stellar/kelp/issues/561)
- use single call to load offers when resetting liabilities, closes [#563](https://github.com/stellar/kelp/issues/563)
- Add missing CLI metrics from inputs (part of [#551](https://github.com/stellar/kelp/issues/551))
- add GOARM versions in metrics, closes [#567](https://github.com/stellar/kelp/issues/567)
- increase default spread in sample config file to avoid op_cross_self errors during submission ([ba35e72a18a793f3fb5241297a87100ff5b6e282](https://github.com/stellar/kelp/commit/ba35e72a18a793f3fb5241297a87100ff5b6e282))
- Refactor volume filter function ([#604](https://github.com/stellar/kelp/issues/604)
- Update README to include steps to install astilectron-bundler ([ccf2bcabc417242dfe3936869f2d8b15853b5cbd](https://github.com/stellar/kelp/commit/ccf2bcabc417242dfe3936869f2d8b15853b5cbd))
- clearly document / revise description of behavior of volume filter in config file and revise tests in dailyVolumeByDate ([#623](https://github.com/stellar/kelp/issues/623))
- clean up root.go basic kelp binary invocation logic ([#568](https://github.com/stellar/kelp/issues/568), [219a557ee5b6b56490cd0aee30d06573e796cc24](https://github.com/stellar/kelp/commit/219a557ee5b6b56490cd0aee30d06573e796cc24))

### Deprecated

- deprecate TICK_INTERVAL_SECONDS in favor of TICK_INTERVAL_MILLIS ([#609](https://github.com/stellar/kelp/issues/609), [2e47abae6749840ef600edf2a0a6316ab66d1137](https://github.com/stellar/kelp/commit/2e47abae6749840ef600edf2a0a6316ab66d1137))

### Fixed

- mirror strategy should ignore backing orders below min volume requirement, closes [#569](https://github.com/stellar/kelp/issues/569)
- move metrics tracker to plugins package to prevent import cycles ([#583](https://github.com/stellar/kelp/issues/583))
- fix DYNAMIC_LDFLAGS ([#587](https://github.com/stellar/kelp/issues/587))
- sample_selltwap.cfg uses incorrect fields (DATA_TYPE_A and DATA_FEED_A_URL), replace them, closes [#598](https://github.com/stellar/kelp/issues/598)
- Add tests for the volume filter (part of [#483](https://github.com/stellar/kelp/issues/483))
- Add test for volume filter function (closes [#483](https://github.com/stellar/kelp/issues/483))
- twap strategy throws error if round returns size near 0, closes [#588](https://github.com/stellar/kelp/issues/588)
- TestMarketID, closes [#594](https://github.com/stellar/kelp/issues/594)
- Rename caps in volume filter tests (part of [#522](https://github.com/stellar/kelp/issues/522))
- add tests for interval time controller ([#605](https://github.com/stellar/kelp/issues/605))
- Validate volume filter config ([#571](https://github.com/stellar/kelp/issues/571))
- Modify tests for volume filter ([d811d406cfa8571aa24504ac85f277e03bb060b3](https://github.com/stellar/kelp/commit/d811d406cfa8571aa24504ac85f277e03bb060b3), [798f548e0845b8eb0272480fc3d314462471212d](https://github.com/stellar/kelp/commit/798f548e0845b8eb0272480fc3d314462471212d), [61e2303670de55d2515caea8a7cd6ae0abee23c3](https://github.com/stellar/kelp/commit/61e2303670de55d2515caea8a7cd6ae0abee23c3), [fa2fed9d7c3d78890c86f8103b5a43bfae2be1af](https://github.com/stellar/kelp/commit/fa2fed9d7c3d78890c86f8103b5a43bfae2be1af), [e41133f00ea26c05123bccc11cac395e23f4b1bc](https://github.com/stellar/kelp/commit/e41133f00ea26c05123bccc11cac395e23f4b1bc), [f909f50677ba1e3511024f1a163ecd7b74f02122](https://github.com/stellar/kelp/commit/f909f50677ba1e3511024f1a163ecd7b74f02122), [df4f2fac5c12bfaf566d9caa631993c430da0b12](https://github.com/stellar/kelp/commit/df4f2fac5c12bfaf566d9caa631993c430da0b12), [56c2d6db2655d38b9d65071eea5b0a7590e0b974](https://github.com/stellar/kelp/commit/56c2d6db2655d38b9d65071eea5b0a7590e0b974), [340d6f16469bd4c4ed8e135a9e3f56ad63a9a6e8](https://github.com/stellar/kelp/commit/340d6f16469bd4c4ed8e135a9e3f56ad63a9a6e8))
- fix botName regex initialization ([554a36b5c22f6fe18d4e7732c92caa49e4ba0ca8](https://github.com/stellar/kelp/commit/554a36b5c22f6fe18d4e7732c92caa49e4ba0ca8))
- spread value in GUI should be correct along with spread % ([#619](https://github.com/stellar/kelp/issues/619))
- bugfix: volumeFilterFn should explicitly take in action buy/sell ([#646](https://github.com/stellar/kelp/issues/646))
- build script should return an error if amplitude key is missing for force releases ([047db942fd7abbfd4ca78fb74ff6d64acc3e2538](https://github.com/stellar/kelp/commit/047db942fd7abbfd4ca78fb74ff6d64acc3e2538))
- build script should return an error if amplitude key is missing for test releases ([89f3d310da58b498689e7ab3faed5a7cc87a2294](https://github.com/stellar/kelp/commit/89f3d310da58b498689e7ab3faed5a7cc87a2294))
- do not crash bot when we encounter a startup event error from Amplitude ([#651](https://github.com/stellar/kelp/issues/651))
- fix priceFeed_test by adjusting upper bound of expected XLM price ([84ac63d76f7fafb87d93724cadaebb75448bfc5e](https://github.com/stellar/kelp/commit/84ac63d76f7fafb87d93724cadaebb75448bfc5e))

## [v1.10.0] - 2020-10-22

### Added

- New Trading Strategy: Pendulum Strategy ([#427](https://github.com/stellar/kelp/issues/427))
- Log value of total assets in a common format ([#433](https://github.com/stellar/kelp/issues/433))
- Always log levels returned for all strategies ([#436](https://github.com/stellar/kelp/issues/436))
- Add code version used to upgrade database in the db_version table ([#445](https://github.com/stellar/kelp/issues/445) and [#447](https://github.com/stellar/kelp/issues/447))
- Database Schema Test Infrastructure also tests indexes on tables ([ad607653c13eb76ba3be5b17a5203a08b2ea11af](https://github.com/stellar/kelp/commit/ad607653c13eb76ba3be5b17a5203a08b2ea11af))
- Upgrade Script for postgres database to include an accountId in the trades database ([#444](https://github.com/stellar/kelp/issues/444))
- Sell TWAP strategy ([#454](https://github.com/stellar/kelp/issues/454))
- Cheaper and more accurate fill tracking (trade tracking) ([#456](https://github.com/stellar/kelp/issues/456) and [a52a1571d39d326640f1eadf4e1ea62e8953a416](https://github.com/stellar/kelp/commit/a52a1571d39d326640f1eadf4e1ea62e8953a416))
- Kelp GUI: Add legal disclaimer about running on mainnet ([#484](https://github.com/stellar/kelp/issues/484))
- Kelp GUI: make pubnet bots available with a single boolean switch ([#475](https://github.com/stellar/kelp/issues/475))
- Add link to PR with new trading template ([#453](https://github.com/stellar/kelp/issues/453))
- Mirror Strategy: track orders triggered on backingExchange by trades on primaryExchange ([#503](https://github.com/stellar/kelp/issues/503))
- Set up CLI data tracking ([#478](https://github.com/stellar/kelp/issues/478), [#516](https://github.com/stellar/kelp/issues/516), [539](https://github.com/stellar/kelp/issues/539), [96bd33d3e6bae5de3bd96361eeb1e195ece55f89](https://github.com/stellar/kelp/commit/96bd33d3e6bae5de3bd96361eeb1e195ece55f89), [84e258208cbae8df6ba6f93e35340351ac58cbba](https://github.com/stellar/kelp/commit/96bd33d3e6bae5de3bd96361eeb1e195ece55f89), [6da8ddc642e9f0112bb20064ada554a7a099f7aa](https://github.com/stellar/kelp/commit/96bd33d3e6bae5de3bd96361eeb1e195ece55f89))
- ccxtExchange should allow fetching binance orderbook with limits between the hardcoded binance limits ([#528](https://github.com/stellar/kelp/issues/528) and [00508f6277f164b58f13fab0ed2e41e9e4241ea7](https://github.com/stellar/kelp/commit/00508f6277f164b58f13fab0ed2e41e9e4241ea7))
- Simple Integration tests in CircleCI ([#51](https://github.com/stellar/kelp/issues/51))

### Changed

- Post-only support for some exchanges when using maker_only submit mode ([#440](https://github.com/stellar/kelp/issues/440))
- update disclaimer in Kelp README ([#485](https://github.com/stellar/kelp/issues/485))
- mirror strategy should handle primary and backing exchange balance calculations and constraints better ([#524](https://github.com/stellar/kelp/issues/524))
- various updates to README ([ee1fe1f847b5ace97272b82de2bf758e4bb732e5](https://github.com/stellar/kelp/commit/ee1fe1f847b5ace97272b82de2bf758e4bb732e5), [c22e46bd46307d31f47829cad0d4f57921abfb9e](https://github.com/stellar/kelp/commit/c22e46bd46307d31f47829cad0d4f57921abfb9e), [2d15399fbf1d0533563d54d0450c3eab950c5525](https://github.com/stellar/kelp/commit/2d15399fbf1d0533563d54d0450c3eab950c5525))
- add winning StellarBattle content to README ([0ba9a2c563a9fb664611f326a03ad8d2e83ccb39](https://github.com/stellar/kelp/commit/0ba9a2c563a9fb664611f326a03ad8d2e83ccb39) and [2baa57f34946f48d4e6592cb7c1832b585a40d19](https://github.com/stellar/kelp/commit/2baa57f34946f48d4e6592cb7c1832b585a40d19))
- add min postgres version number to README ([#514](https://github.com/stellar/kelp/issues/514))
- Fix sample config files ([#538](https://github.com/stellar/kelp/issues/538))
- mirror strategy should allow different divide by values for bid and ask sides, deprecate VOLUME_DIVIDE_BY config field ([#545](https://github.com/stellar/kelp/issues/545))
- mark Kraken exchange as tested ([2e3db3c6d530663af783604c59ba0c7407b9ba7d](https://github.com/stellar/kelp/commit/2e3db3c6d530663af783604c59ba0c7407b9ba7d))

### Deprecated

- mirror strategy should allow different divide by values for bid and ask sides, deprecate VOLUME_DIVIDE_BY config field ([#545](https://github.com/stellar/kelp/issues/545))

### Fixed

- Kelp GUI: fix issue of fiat currency dropdown not updating correctly ([fe19dcbaac0845e5bec7415528ffe02db93245af](https://github.com/stellar/kelp/commit/fe19dcbaac0845e5bec7415528ffe02db93245af))
- fix index out of range when getting prices from sdex ([#416](https://github.com/stellar/kelp/issues/416))
- Fix baseAmount used when placing orders ([#435](https://github.com/stellar/kelp/issues/435))
- Fix FetchTrades for Kraken ([#450](https://github.com/stellar/kelp/issues/450))
- sellSideStrategy.go#PreUpdate does not call GetLevels when base asset is 0.0 ([#457](https://github.com/stellar/kelp/issues/457))
- KrakenExchange should get latest cursor in seconds instead of millis ([#465](https://github.com/stellar/kelp/issues/465))
- bot should crash if delete cycles threshold is exceeded ([#471](https://github.com/stellar/kelp/issues/471))
- remove minOrderSizeBase from UUID in sellTwapLevelProvider.go ([#482](https://github.com/stellar/kelp/issues/482))
- Kelp GUI: fix another instance of OSPath.String() being called ([#430](https://github.com/stellar/kelp/issues/430))
- failure to submit ops (async or sync) should count towards the delete cycles threshold ([#498](https://github.com/stellar/kelp/issues/498))
- mirror strategy should prepend deleteOps before both bid and ask ops ([#501](https://github.com/stellar/kelp/issues/501))
- mirror strategy: log num trades received from backing exchange on triggered fill ([#505](https://github.com/stellar/kelp/issues/505))
- Kelp GUI: Propagate bot initialization & startup errors back to GUI ([#506](https://github.com/stellar/kelp/issues/506))
- More granular Kelp AppNames ([#488](https://github.com/stellar/kelp/issues/488))
- Kelp GUI: disallow invalid characters in bot name ([#429](https://github.com/stellar/kelp/issues/429))
- mirror strategy fails to start up without db enabled, nil pointer dereference ([#525](https://github.com/stellar/kelp/issues/525))
- modify offers in mirror strategy is not correctly adjusting price and amount ([#526](https://github.com/stellar/kelp/issues/526))
- Rounding issues in mirror strategy causing offers to not be placed ([#541](https://github.com/stellar/kelp/issues/541))

## [v1.9.0] - 2020-05-07

### Added

- Kelp GUI: allows revealing the log file on startup ([bbd709736f25fdf8f4809fb5e4f017ce1d405afa](https://github.com/stellar/kelp/commit/bbd709736f25fdf8f4809fb5e4f017ce1d405afa))
- Kelp GUI: start backend server before sending ready string message ([f2d75d52bfb9dbbee74414b863b813a204f85a53](https://github.com/stellar/kelp/commit/f2d75d52bfb9dbbee74414b863b813a204f85a53))
- Kelp GUI: explicit quit button inside the Kelp GUI window ([846fcf0517be36d3143967e43f33d2a2238aa719](https://github.com/stellar/kelp/commit/846fcf0517be36d3143967e43f33d2a2238aa719))
- update build script to introduce new force option (-f, --force) ([f49f8778f226dbe0069c057068a7557d2411e955](https://github.com/stellar/kelp/commit/f49f8778f226dbe0069c057068a7557d2411e955))
- Kelp GUI: improve ccxt download to show progress for better visibility ([9a9fad13701b84a54c5441175328eae457a2c454](https://github.com/stellar/kelp/commit/9a9fad13701b84a54c5441175328eae457a2c454))
- Kelp GUI: copy ccxt zip from source near binary to dotKelp/ccxt folder ([8ea4b1e5b8c6438d67daea24c5311a3d80fb1ac9](https://github.com/stellar/kelp/commit/8ea4b1e5b8c6438d67daea24c5311a3d80fb1ac9))
- Kelp GUI: allow copy paste keyboard shortcuts and add to menu bar along with Quit option ([3dfb14a3dffcea2ddeb474501cc24185158f0932](https://github.com/stellar/kelp/commit/3dfb14a3dffcea2ddeb474501cc24185158f0932))
- improve build script process used to package Kelp GUI ([35533687b828880f19b33d953ba21790f9e86414](https://github.com/stellar/kelp/commit/35533687b828880f19b33d953ba21790f9e86414))
- add image logo to README ([61bbf654dd8ead5dc76c86da1b5c637b9780af5a](https://github.com/stellar/kelp/commit/61bbf654dd8ead5dc76c86da1b5c637b9780af5a))
- Kelp GUI: add reload button in menu ([64807a4538d19710dca2a91b907c36334da10f7a](https://github.com/stellar/kelp/commit/64807a4538d19710dca2a91b907c36334da10f7a))

### Changed

- Kelp GUI: allow bots to be deleted when in Initializing state ([1491b61ddfd9f12c76faa91eeff5cd620c508c64](https://github.com/stellar/kelp/commit/1491b61ddfd9f12c76faa91eeff5cd620c508c64))
- Kelp GUI: fix generation of bundler.json and bind files to reduce redundant astilectron builds ([5a5967c78e087023ac96bf64e5683a52ff5af7ae](https://github.com/stellar/kelp/commit/5a5967c78e087023ac96bf64e5683a52ff5af7ae))
- Kelp GUI: package for windows via tail file web server with cors-enabled /ping endpoint and add --no-electron ([71729111be7bc23eb9aac601b9dd407aa607503d](https://github.com/stellar/kelp/commit/71729111be7bc23eb9aac601b9dd407aa607503d))
- Kelp GUI: Fix filepaths for windows by introducing the kelpos.OSPath ([e6d89f7a79774b6e36c79f13dc735d7af9216dbd](https://github.com/stellar/kelp/commit/e6d89f7a79774b6e36c79f13dc735d7af9216dbd))
- Kelp GUI: Fix basepath and use pingURL to ping server in tailFile before redirect ([20a5cf4c08491aa49ecc385c643237d8adec5a9c](https://github.com/stellar/kelp/commit/20a5cf4c08491aa49ecc385c643237d8adec5a9c))
- Kelp GUI: do not attempt to add trustlines for assets created by trading account ([9fd7afc5ad03ae5a1473dd6dedc0e319bdddaea1](https://github.com/stellar/kelp/commit/9fd7afc5ad03ae5a1473dd6dedc0e319bdddaea1))
- Kelp GUI: do not allow trading from the issuer account ([2081ad9f83a6a69623150115446e586d8baa1108](https://github.com/stellar/kelp/commit/2081ad9f83a6a69623150115446e586d8baa1108))
- Upgrade stellar SDK dependency to 'horizonclient-v3.0.0' tag from stellar/go to make it compatible with protocol 13 ([9884a0402d3f0d7307f6de03d46013a3605c79bf](https://github.com/stellar/kelp/commit/9884a0402d3f0d7307f6de03d46013a3605c79bf))
- when setting fee if endpoint is not available (eg: custom network) then use maxOpFeeStroops ([5e8085c790214fa3569ed5c76ca622e8da8e4834](https://github.com/stellar/kelp/commit/5e8085c790214fa3569ed5c76ca622e8da8e4834))
- add back missing dependencies of asticode to glide.lock ([a01c0b21da7e7a265a0a4a54130c8f135a1a18cb](https://github.com/stellar/kelp/commit/a01c0b21da7e7a265a0a4a54130c8f135a1a18cb))
- Kelp GUI: windows does not use zip version of ccxt-rest binary because 'unzip' is not supported by default in WSL ([f86d17ea056dbdb9f65f19fc7dda115574cf9e5d](https://github.com/stellar/kelp/commit/f86d17ea056dbdb9f65f19fc7dda115574cf9e5d))
- Kelp GUI: autogenerated and empty bots run once every minute ([6e853bfcbd64783ea49397c7616d1a81a3839c0c](https://github.com/stellar/kelp/commit/6e853bfcbd64783ea49397c7616d1a81a3839c0c))
- Kelp GUI: adjust frontend timer intervals and add comments ([6fb8d428b59a2e517d8f952837169b416b9a45c6](https://github.com/stellar/kelp/commit/6fb8d428b59a2e517d8f952837169b416b9a45c6))
- Kelp GUI: automatically pay test accounts with 1000 units of any asset issued by GBMMZ... public issuer ([3c87298c8b4ce88045d542b872f825532071fc86](https://github.com/stellar/kelp/commit/3c87298c8b4ce88045d542b872f825532071fc86))

### Fixed

- Kelp GUI: fix too many open files error ([df4cb9e87eee5c537f22cd19b49a56c0709c610d](https://github.com/stellar/kelp/commit/df4cb9e87eee5c537f22cd19b49a56c0709c610d))
- Kelp GUI: fix windows bash execution ([ec74f1b4ad019864b79a8f35e4638947bebe621f](https://github.com/stellar/kelp/commit/ec74f1b4ad019864b79a8f35e4638947bebe621f))
- Kelp GUI: run ccxt on windows ([9f9ab964a299cd1c950838d2cb1c4bd32bb426ba](https://github.com/stellar/kelp/commit/9f9ab964a299cd1c950838d2cb1c4bd32bb426ba))
- Kelp GUI: use github.com/pkg/browser to call cross-platform OpenURL function ([8729754b4df07e2c0db537780a1dd41c90e44b09](https://github.com/stellar/kelp/commit/8729754b4df07e2c0db537780a1dd41c90e44b09))
- Kelp GUI: default windows to --no-electron in bat file ([6a6d9ccae71b788dee55970f072edcc8f2c5a80d](https://github.com/stellar/kelp/commit/6a6d9ccae71b788dee55970f072edcc8f2c5a80d))
- Kelp GUI: refactor os paths used to accommodate for 260 character file path limit in windows ([ae33d83c7ba2344e085816f4071b0763a4a1336e](https://github.com/stellar/kelp/commit/ae33d83c7ba2344e085816f4071b0763a4a1336e))

## [v1.8.1] - 2020-02-17

### Changed

- throw error on startup when FILL_TRACKER_SLEEP_MILLIS is not set but POSTGRES_DB is set ([99799da07bd2afafc061d410f9fa72b0b0332d75](https://github.com/stellar/kelp/commit/99799da07bd2afafc061d410f9fa72b0b0332d75))
- log offer if isSelling check fails ([ea505bdfd6c41ecf71eb1b13e3c5f0c1cb7666a3](https://github.com/stellar/kelp/commit/ea505bdfd6c41ecf71eb1b13e3c5f0c1cb7666a3))

### Fixed

- upgrade horizonclient to patched version to fix load offers issues ([c0a4e3ac4eadfd34b31081ff911169a37d26d7d5](https://github.com/stellar/kelp/commit/c0a4e3ac4eadfd34b31081ff911169a37d26d7d5))
- upgrade horizonclient to patched version to fix delete offer op issue ([2cbfb6782915f1bfd949aa2821d94e9bbf735d6f](https://github.com/stellar/kelp/2cbfb6782915f1bfd949aa2821d94e9bbf735d6f))
- workaround empty trades error ([f6d31c2587c3bc7fb4a68a7e7ed6c1777021f001](https://github.com/stellar/kelp/commit/f6d31c2587c3bc7fb4a68a7e7ed6c1777021f001))
- update priceFeed_test#wantUpperBoundXLM ([a8dbcf4afbdb896e7fe5fa85de611a48b3112db9](https://github.com/stellar/kelp/commit/a8dbcf4afbdb896e7fe5fa85de611a48b3112db9))

## [v1.8.0] - 2020-02-11

### Added

- Dockerize Kelp binary ([b61207012dd10b44220acf644703aa346834778c](https://github.com/stellar/kelp/commit/b61207012dd10b44220acf644703aa346834778c))
- Kelp UI: Wrap GUI as a standalone desktop application using Electron ([b725cbaf9c67e8d3b9aea29316c5ec22d168c81e](https://github.com/stellar/kelp/commit/b725cbaf9c67e8d3b9aea29316c5ec22d168c81e))
- combine build and test tasks in circleci ([b725cbaf9c67e8d3b9aea29316c5ec22d168c81e](https://github.com/stellar/kelp/commit/b725cbaf9c67e8d3b9aea29316c5ec22d168c81e))
- script to build pre-compiled binaries for CCXT-rest ([b0d608f092b7dd461ec14b350c5e6d4789c7fa01](https://github.com/stellar/kelp/commit/b0d608f092b7dd461ec14b350c5e6d4789c7fa01))
- add support for dynamic headers in CCXT for exchanges such as coinbase ([335d191e6d5b4cadc738454023eb65450a008d8b](https://github.com/stellar/kelp/commit/335d191e6d5b4cadc738454023eb65450a008d8b))
- allow bot to write its trades to a postgres SQL database via a config param ([493b4b004c7363634141723e40350dae0edb9fad](https://github.com/stellar/kelp/commit/493b4b004c7363634141723e40350dae0edb9fad), [a6ffc8c770b03999c58fd2f589b58622fa80ac00](https://github.com/stellar/kelp/commit/a6ffc8c770b03999c58fd2f589b58622fa80ac00))
- new filter system for risk management along with a set of some basic filters: 'volume', 'price', 'priceFeed' ([11d4927770b2fbbade2dc8f61055f4faa504af17](https://github.com/stellar/kelp/commit/11d4927770b2fbbade2dc8f61055f4faa504af17), [66ea6105938434c090b28d3b7cb65d32d5100a62](https://github.com/stellar/kelp/commit/66ea6105938434c090b28d3b7cb65d32d5100a62), [3e0c240c7c1aabd618f45634589237c3dcd91cd3](https://github.com/stellar/kelp/commit/3e0c240c7c1aabd618f45634589237c3dcd91cd3), [9062d7d01904990ff8690932f3702023b27e491e](https://github.com/stellar/kelp/commit/9062d7d01904990ff8690932f3702023b27e491e), [391d3fbcc20e53e4daf556db4617f59e0f9a98e9](https://github.com/stellar/kelp/commit/391d3fbcc20e53e4daf556db4617f59e0f9a98e9))
- modifiers to price feed: 'mid', 'ask', 'bid', 'last' ([116f7d1c1762b23c93389d13120e37111a3d6ef7](https://github.com/stellar/kelp/commit/116f7d1c1762b23c93389d13120e37111a3d6ef7), [afb56289b86cf5412580a0b0536b8230e3a3a37c](https://github.com/stellar/kelp/commit/afb56289b86cf5412580a0b0536b8230e3a3a37c))
- new type of 'function' price feed with the following functions out-of-the-box: 'max', 'invert' ([412b81cdf925b4d2c498a8d691e86411f3ba6b4a](https://github.com/stellar/kelp/commit/412b81cdf925b4d2c498a8d691e86411f3ba6b4a))
- allow custom starting point from where to load trades into db using FILL_TRACKER_LAST_TRADE_CURSOR_OVERRIDE config param ([4c19915f4795732c75a76eeff07160be29f426d6](https://github.com/stellar/kelp/commit/4c19915f4795732c75a76eeff07160be29f426d6))
- Kelp GUI should use pre-compiled CCXT binary to expand access to exchanges ([fba752f99fff79a10a2a308efb6794b251ff0d03](https://github.com/stellar/kelp/commit/fba752f99fff79a10a2a308efb6794b251ff0d03))
- UI feedback during Kelp GUI app startup ([12eccd2a566e68a707e5777d9b2759c239f10cb5](https://github.com/stellar/kelp/commit/12eccd2a566e68a707e5777d9b2759c239f10cb5))
- Kelp GUI should have a different version number from the Kelp CLI ([ba297f6d6c0f21da12a93f373bbcf16868d86958](https://github.com/stellar/kelp/commit/ba297f6d6c0f21da12a93f373bbcf16868d86958))
- Kelp GUI: Parallelize loading of CCXT instances ([0aa5700c75eb4cc1af2b817eef3961ad6aef63f7](https://github.com/stellar/kelp/commit/0aa5700c75eb4cc1af2b817eef3961ad6aef63f7))

### Changed

- Upgraded horizon types to use hProtocol package ([4af564dd9aeeb976685e381470f8a9fa0626b49e](https://github.com/stellar/kelp/commit/4af564dd9aeeb976685e381470f8a9fa0626b49e))
- Upgrade horizonclient to v2 to support API of horizon v1 API ([ba198426b99e7919a16ec998503ec5d0143d38bf](https://github.com/stellar/kelp/commit/ba198426b99e7919a16ec998503ec5d0143d38bf))
- Upgraded Go SDK to use horizonclient package ([585080c76f173acd5a1348f3f662796d5aeda719](https://github.com/stellar/kelp/commit/585080c76f173acd5a1348f3f662796d5aeda719))
- Upgraded Go SDK usage to `txnbuild` package instead of `build` package ([c18c97f388d3a605b9c48edb5085008791467a1c](https://github.com/stellar/kelp/commit/c18c97f388d3a605b9c48edb5085008791467a1c))
- Multiple usability improvements to the Kelp UI ([f7db6c8430c834040020efa7c58ed860ff303abc through f765ae00d73f4a6a3d6eedf35de6d5528a5f455f](https://github.com/stellar/kelp/compare/f7db6c8430c834040020efa7c58ed860ff303abc~1...f765ae00d73f4a6a3d6eedf35de6d5528a5f455f))
- Guarantee fixed number of successful runs of update cycle via the `--iter` cli param ([4845a6220a5091cd97c6833c359077c7a3afc291](https://github.com/stellar/kelp/commit/4845a6220a5091cd97c6833c359077c7a3afc291))
- updated README ([c58982c25e8ead8c91ca17f09d4c96cc3705d772](https://github.com/stellar/kelp/commit/c58982c25e8ead8c91ca17f09d4c96cc3705d772), [2cb57326f37b7f68ed9d58710eaca4fec0111113](https://github.com/stellar/kelp/commit/2cb57326f37b7f68ed9d58710eaca4fec0111113), [dddd9707b6c20e259595979fb96c8b95eb634757](https://github.com/stellar/kelp/commit/dddd9707b6c20e259595979fb96c8b95eb634757))
- current official support for only go1.13 ([45d80334c5772a139c1066731d5937977e590fee](https://github.com/stellar/kelp/commit/45d80334c5772a139c1066731d5937977e590fee))

### Deprecated

### Removed

- remove travis.yml config file ([54d4fc88e83e6f5f3211226cadd84118e9142995](https://github.com/stellar/kelp/commit/54d4fc88e83e6f5f3211226cadd84118e9142995))

### Fixed

- git bash windows compatibility, replace `rev` command in build.sh and clean.sh ([8ea336c379e6770a7ee4646aa3750ca51ed6f203](https://github.com/stellar/kelp/commit/8ea336c379e6770a7ee4646aa3750ca51ed6f203))
- improve number.AsRatio() conversion using stellar/go/price for more accurate pricing on centralized exchanges ([59cabd6bf81e61a237f33c25e319530937941d76](https://github.com/stellar/kelp/commit/59cabd6bf81e61a237f33c25e319530937941d76))
- fix number.AsString() method ([f942cf8f24f65d54e9aa9c232594c73fef236e5f](https://github.com/stellar/kelp/commit/f942cf8f24f65d54e9aa9c232594c73fef236e5f))
- fix cursor and cost param in trade parsing and ccxtExchange_test ([b6eb0411aa8f01dd1c4acd07d28d2886f75bfc49](https://github.com/stellar/kelp/commit/b6eb0411aa8f01dd1c4acd07d28d2886f75bfc49))
- return an error when loading existing offers fails instead of ignoring ([95503d943d1152c6524d1fae5efde762adbaf9a6](https://github.com/stellar/kelp/commit/95503d943d1152c6524d1fae5efde762adbaf9a6))
- ccxtExchange should correctly populate cost of trade ([db4531d5866853681d3dc71c222e42b0416c044d](https://github.com/stellar/kelp/commit/db4531d5866853681d3dc71c222e42b0416c044d))

### Security

- upgrade yarn.lock js dependencies to address any security concerns in js dependency libraries used by Kelp GUI ([a6c03336f0d9f1bc6874eabe2887171a4dd4a369](https://github.com/stellar/kelp/commit/a6c03336f0d9f1bc6874eabe2887171a4dd4a369), [77ec937e2175082969aae7b133daf9ea0cf9a350](https://github.com/stellar/kelp/commit/77ec937e2175082969aae7b133daf9ea0cf9a350))

## [v1.7.2] - 2019-08-26

### Added

- add triggers to modification log line in sdex ([a9991dcfde025c20239bf28f35c0f28d73b1107c](https://github.com/stellar/kelp/commit/a9991dcfde025c20239bf28f35c0f28d73b1107c))

### Fixed

- fix fill tracker to also work with path_payment type operations for sdex, fixes #219 ([fb6f4d41bb395769fdc4965e9c0d76033bbcd192](https://github.com/stellar/kelp/commit/fb6f4d41bb395769fdc4965e9c0d76033bbcd192))
- fix op_underfunded error when replenishing top offer ([b56d7b51b467a4c2c0f059c8620b59800049c321](https://github.com/stellar/kelp/commit/b56d7b51b467a4c2c0f059c8620b59800049c321))
- new oversell trigger during modify to check need to reduce amount of existing offers ([82285f3c5b4dcb6c05d5e400a49d57e23c7942b3](https://github.com/stellar/kelp/commit/82285f3c5b4dcb6c05d5e400a49d57e23c7942b3))

## [v1.7.1] - 2019-07-18

### Fixed

- Fill tracker order action now correctly reflects direction of trade ([5392d3e89787cf39993abaa02f5eea69be6e8c59](https://github.com/stellar/kelp/commit/5392d3e89787cf39993abaa02f5eea69be6e8c59))

## [v1.7.0] - 2019-05-05

### Added

- Bundle React.js into Kelp with 3 modes of running with 'server' command in dev mode ([857214c14a5fd2c67a20d618f45614e33c6a97ae](https://github.com/stellar/kelp/commit/857214c14a5fd2c67a20d618f45614e33c6a97ae))
- added UI components for GUI ([22e3d4242e93f2f0ddaf2e66a3f796ed668e1a0e](https://github.com/stellar/kelp/commit/22e3d4242e93f2f0ddaf2e66a3f796ed668e1a0e))
- allow setting custom URL for CCXT-rest server ([3c9af0cb098e0476512de2d002693b9b6cae426b](https://github.com/stellar/kelp/commit/3c9af0cb098e0476512de2d002693b9b6cae426b))

### Changed

- improve function number.AsRatio() for more accurate pricing on centralized exchanges ([5d33101192a4834ee186228c4c3a17d6112086d7](https://github.com/stellar/kelp/commit/5d33101192a4834ee186228c4c3a17d6112086d7), [f193eeef1c7cc328a529e6403af67f0f325ca0e6](https://github.com/stellar/kelp/commit/f193eeef1c7cc328a529e6403af67f0f325ca0e6))
- increased limit when checking for open offers, reducing total requests to horizon ([5416d78c754360144ed1f7e3cc0b31135eb47fea](https://github.com/stellar/kelp/commit/5416d78c754360144ed1f7e3cc0b31135eb47fea))

### Fixed

- Use v0.0.4 for ccxt-rest to fix travis build and update the instructions in our README file. The APIs have diverged with the latest release of ccxt-rest v1.0.0 so we are sticking to the older version for now ([659bb20560a018c766c5c5db1bed55df922b7a2e](https://github.com/stellar/kelp/commit/659bb20560a018c766c5c5db1bed55df922b7a2e)).

## [v1.6.1] - 2019-04-12

### Added

- Use app name and version headers from horizon client in Go SDKs ([d21a75fcdbff323de46e3d2a46f37d64831b1cb7](https://github.com/stellar/kelp/commit/d21a75fcdbff323de46e3d2a46f37d64831b1cb7))
- Add overrides for remaining orderConstraints via configs, fixing precision errors on centralized exchanges ([6d989c8d78ea7c088e36b2b6f2bb7679013617d0](https://github.com/stellar/kelp/commit/6d989c8d78ea7c088e36b2b6f2bb7679013617d0))
- Add support for passing in params and headers to CCXT exchanges, such as coinbasepro ([e7c76fe14191f14410aa0bcc34b06e202cc1c020](https://github.com/stellar/kelp/commit/e7c76fe14191f14410aa0bcc34b06e202cc1c020))

### Fixed

- Fixed "Asset converter could not recognize string" error when trading using krakenExchange ([258f1d67d3899ed21c3ee69cefd6299ea1c8bd4a](https://github.com/stellar/kelp/commit/258f1d67d3899ed21c3ee69cefd6299ea1c8bd4a))
- Turn off minBaseVolume checks in mirror strategy when OFFSET_TRADES=false ([82e58b49381973efa5adc12c0f0238432f6cce2c](https://github.com/stellar/kelp/commit/82e58b49381973efa5adc12c0f0238432f6cce2c))
- fix monitoring by only checking google auth when related config values are passed in ([860d76b0c089efa62299093ff9ccf2d7b868a14c](https://github.com/stellar/kelp/commit/860d76b0c089efa62299093ff9ccf2d7b868a14c))

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

[Unreleased]: https://github.com/stellar/kelp/compare/v1.11.0...HEAD
[v1.11.0]: https://github.com/stellar/kelp/compare/v1.10.0...v1.11.0
[v1.10.0]: https://github.com/stellar/kelp/compare/v1.9.0...v1.10.0
[v1.9.0]: https://github.com/stellar/kelp/compare/v1.8.1...v1.9.0
[v1.8.1]: https://github.com/stellar/kelp/compare/v1.8.0...v1.8.1
[v1.8.0]: https://github.com/stellar/kelp/compare/v1.7.2...v1.8.0
[v1.7.2]: https://github.com/stellar/kelp/compare/v1.7.1...v1.7.2
[v1.7.1]: https://github.com/stellar/kelp/compare/v1.7.0...v1.7.1
[v1.7.0]: https://github.com/stellar/kelp/compare/v1.6.1...v1.7.0
[v1.6.1]: https://github.com/stellar/kelp/compare/v1.6.0...v1.6.1
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
