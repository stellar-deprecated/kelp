# Kelp

[![GitHub last commit](https://img.shields.io/github/last-commit/lightyeario/kelp.svg?style=for-the-badge)][github-last-commit]
[![Github All Releases](https://img.shields.io/github/downloads/lightyeario/kelp/total.svg?style=for-the-badge)][github-releases]
[![license](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge&longCache=true)][license-apache]

[![Build Status](https://travis-ci.com/lightyeario/kelp.svg?branch=master)](https://travis-ci.com/lightyeario/kelp)
[![GitHub issues](https://img.shields.io/github/issues/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-issues]
[![GitHub closed issues](https://img.shields.io/github/issues-closed/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-issues-closed]
[![GitHub pull requests](https://img.shields.io/github/issues-pr/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-pulls]
[![GitHub closed pull requests](https://img.shields.io/github/issues-pr-closed/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-pulls-closed]

Kelp is a free and open-source trading bot for the [Stellar universal marketplace][stellarx].
Kelp is still in beta, so please submit any issues ([bug reports][github-bug-report] and [feature requests][github-feature-request]).

Kelp includes several configurable trading strategies, and its modular design allows you to customize your algorithms, exchange integrations, and assets. You can define your own parameters or use the existing repository to quickly implement a trading bot. With Kelp, you could be up and trading on Stellar in a matter of minutes.

Kelp is pre-configured to:

- Make spreads and make markets
- Create liquidity and facilitate price-discovery for ICOs
- Price and trade custom [stablecoins][stablecoin]
- Mimic order books from other exchanges

To learn more about the Stellar protocol check out [this video created by Lumenauts][sdex explainer video]. You can also search [Stellar's Q&A][stellar q&a].

# Table of Contents

   * [Getting Started](#getting-started)
      * [Download Binary](#download-binary)
      * [Compile from Source](#compile-from-source)
      * [Running Kelp](#running-kelp)
      * [Using CCXT](#using-ccxt)
      * [Be Smart and Go Slow](#be-smart-and-go-slow)
   * [Components](#components)
      * [Strategies](#strategies)
      * [Price Feeds](#price-feeds)
      * [Configuration Files](#configuration-files)
      * [Exchanges](#exchanges)
      * [Plugins](#plugins)
      * [Directory Structure](#directory-structure)
      * [Accounting](#accounting)
   * [Examples](#examples)
      * [Walkthrough Guides](#walkthrough-guides)
      * [Configuration Files](#configuration-files-1)
   * [Contributing](#contributing)
   * [Changelog](#changelog)
   * [Questions &amp; Improvements](#questions--improvements)

# Getting Started

To get started with Kelp, either download the pre-compiled binary for your platform from the [Github Releases Page][github-releases] or [compile Kelp from source](#compile-from-source).

There is **one** binary associated with this project: `kelp`. Once the binary is downloaded, run the bot by following the instructions in [Running Kelp](#running-kelp).

## Download Binary

You can find the pre-compiled binary for your platform from the [Github Releases Page][github-releases].

Here is a list of binaries for the most recent release **v1.1.0**:

| Platform       | Architecture | Binary File Name |
| -------------- | ------------ | ---------------- |
| MacOS (Darwin) | 64-bit       | [kelp-v1.1.0-darwin-amd64.tar](https://github.com/lightyeario/kelp/releases/download/v1.1.0/kelp-v1.1.0-darwin-amd64.tar) |
| Windows        | 64-bit       | [kelp-v1.1.0-windows-amd64.tar](https://github.com/lightyeario/kelp/releases/download/v1.1.0/kelp-v1.1.0-windows-amd64.tar) |
| Linux          | 64-bit       | [kelp-v1.1.0-linux-amd64.tar](https://github.com/lightyeario/kelp/releases/download/v1.1.0/kelp-v1.1.0-linux-amd64.tar) |
| Linux          | 64-bit arm   | [kelp-v1.1.0-linux-arm64.tar](https://github.com/lightyeario/kelp/releases/download/v1.1.0/kelp-v1.1.0-linux-arm64.tar) |
| Linux          | 32-bit arm5  | [kelp-v1.1.0-linux-arm5.tar](https://github.com/lightyeario/kelp/releases/download/v1.1.0/kelp-v1.1.0-linux-arm5.tar) |
| Linux          | 32-bit arm6  | [kelp-v1.1.0-linux-arm6.tar](https://github.com/lightyeario/kelp/releases/download/v1.1.0/kelp-v1.1.0-linux-arm6.tar) |
| Linux          | 32-bit arm7  | [kelp-v1.1.0-linux-arm7.tar](https://github.com/lightyeario/kelp/releases/download/v1.1.0/kelp-v1.1.0-linux-arm7.tar) |

After you _untar_ the downloaded file, change to the generated directory (`kelp-v1.1.0`) and invoke the `kelp` binary.

Here's an example to get you started (replace `filename` with the name of the file that you download):

    tar xvf filename
    cd kelp-v1.1.0
    ./kelp

To run the bot in simulation mode, try this command:

    ./kelp trade -c sample_trader.cfg -s buysell -f sample_buysell.cfg --sim

## Compile from Source

_Note for Windows Users: You should use a [Bash Shell][bash] to follow the steps below. This will give you a UNIX environment in which to run your commands and will enable the `./scripts/build.sh` bash script to work correctly._

To compile Kelp from source:

1. [Download][golang-download] and [setup][golang-setup] Golang.
2. [Install Glide][glide-install] for dependency management
    * `curl https://glide.sh/get | sh`
3. Clone the repo into `$GOPATH/src/github.com/lightyeario/kelp`:
    * `git clone git@github.com:lightyeario/kelp.git`
4. Change to the kelp directory and install the dependencies:
    * `glide install`
5. Build the binaries using the provided build script (the _go install_ command will produce a faulty binary):
    * `./scripts/build.sh`
6. Confirm one new binary file:
    * `./bin/kelp`
7. Set up CCXT to use an expanded set of priceFeeds or orderbooks (see the [Using CCXT](#using-ccxt) section for details)

## Running Kelp

Kelp places orders on the [Stellar marketplace][stellarx] based on the selected strategy. Configuration files specify the Stellar account and strategy details.

These are the following commands available from the `kelp` binary:
- `exchanges`: Lists the available exchange integrations
- `strategies`: Lists the available strategies
- `trade`: Trades with a specific strategy against the Stellar universal marketplace
- `version`: Version and build information
- `help`: Help about any command

The `trade` command has three parameters which are:

- **botConf**: full path to the _.cfg_ file with the account details, [sample file here](examples/configs/trader/sample_trader.cfg).
- **strategy**: the strategy you want to run (_sell_, _buysell_, _balanced_, _mirror_, _delete_).
- **stratConf**: full path to the _.cfg_ file specific to your chosen strategy, [sample files here](examples/configs/trader/).

Here's an example of how to start the trading bot with the _buysell_ strategy:

`kelp trade --botConf ./path/trader.cfg --strategy buysell --stratConf ./path/buysell.cfg`

If you are ever stuck, just invoke the `kelp` binary directly or type `kelp help [command]` for help with a specific command.

## Using CCXT

You can use the [CCXT][ccxt] library via the [CCXT REST API Wrapper][ccxt-rest] to fetch prices and orderbooks from a larger number of exchanges.

You will need to run the CCXT REST server on `localhost:3000` so Kelp can connect to it. In order to run CCXT you should install [docker][docker] (linux: `sudo apt install -y docker.io`) and run the CCXT-REST docker image configured to port `3000` (linux: `sudo docker run -p 3000:3000 -d franzsee/ccxt-rest`). You can find more details on the [CCXT_REST github page][ccxt-rest]. The CCXT-REST server **must** be running on port `3000` before you start up the Kelp bot.

You can list the exchanges (`./kelp exchanges`) to get the full list of supported exchanges via CCXT.

_Note: this integration is still **experimental** and is also **incomplete**. Please use at your own risk._

## Be Smart and Go Slow

_Whenever you trade on Stellar, you are trading with volatile assets, in volatile markets, and you risk losing money. Use Kelp at your own risk. There is no guarantee you'll make a profit from using our bots or strategies. In fact, if you set bad parameters or market conditions change, Kelp might help you **lose** money very fast. So be smart and go slow._

# Components

Kelp includes an assortment of strategies, price feeds, and plugins you can use to customize your bot. Kelp also enables you to create your own trading strategies.

## Strategies

Strategies are at the core of Kelp. Without them it's just lazy, capable of nothing, thinking of nothing, doing nothing, like our friend [scooter][scooter video] here. The strategies give your bot purpose. Each approaches the market in a different way and is designed to achieve a particular goal.

The following strategies are available **out of the box** with Kelp:

- sell ([source](plugins/sellStrategy.go)):

    - **What:** creates sell offers based on a reference price with a pre-specified liquidity depth
    - **Why:** To sell tokens at a fixed price or at a price that changes based on an external reference price
    - **Who:** An issuer could use Sell to distribute tokens from an ICO pre-sale
    - **Complexity**: Beginner

- buysell ([source](plugins/buysellStrategy.go)):

    - **What:** creates buy and sell offers based on a specific reference price and a pre-specified liquidity depth while maintaining a [spread][spread].
    - **Why:** To make the market for tokens based on a fixed or external reference price.
    - **Who:** Anyone who wants to create liquidity for a stablecoin or [fiat][fiat] token
    - **Complexity:** Beginner

- balanced ([source](plugins/balancedStrategy.go)):
    - **What:** dynamically prices two tokens based on their relative demand. For example, if more traders buy token A _from_ the bot (the traders are therefore selling token B), the bot will automatically raise the price for token A and drop the price for token B.
    - **Why:** To let the market surface the _true price_ for one token in terms of another.
    - **Who:** Market makers and traders for tokens that trade only on Stellar 
    - **Complexity:** Intermediate

- mirror ([source](plugins/mirrorStrategy.go)):

    - **What:** mirrors an orderbook from another exchange by placing the same orders on Stellar after including a [spread][spread]. _Note: covering your trades on the backing exchange is not currently supported out-of-the-box_.
    - **Why:** To [hedge][hedge] your position on another exchange whenever a trade is executed to reduce inventory risk while keeping a spread
    - **Who:** Anyone who wants to reduce inventory risk and also has the capacity to take on a higher operational overhead in maintaining the bot system.
    - **Complexity:** Advanced

- delete ([source](plugins/deleteStrategy.go)):

    - **What:** deletes your offers from both sides of the specified orderbook. _Note: does not need a strategy-specific config file_.
    - **Why:** To kill the offers placed by the bot. _This is not a trading strategy but is used for operational purposes only_.
    - **Who:** Anyone managing the operations of the bot who wants to stop all activity by the bot.
    - **Complexity:** Beginner

## Price Feeds

Price Feeds fetch the price of an asset from an external source. The following price feeds are available **out of the box** with Kelp:

- coinmarketcap: fetches the price of tokens from [CoinMarketCap][cmc]
- fiat: fetches the price of a [fiat][fiat] currency from the [CurrencyLayer API][currencylayer]
- exchange: fetches the price from an exchange you specify, such as Kraken or Poloniex. You can also use the [CCXT][ccxt] integration to fetch prices from a wider range of exchanges (see the [Using CCXT](#using-ccxt) section for details)
- fixed: sets the price to a constant

## Configuration Files

Each strategy you implement needs a configuration file. The format of the configuration file is specific to the selected strategy. You can use these files to customize parameters for your chosen strategy.

For more details, check out the [examples section](#configuration-files-1) of the readme.

## Exchanges

Exchange integrations provide data to trading strategies and allow you to [hedge][hedge] your positions on different exchanges. The following [exchange integrations](plugins) are available **out of the box** with Kelp:

- sdex ([source](plugins/sdex.go)): The [Stellar Decentralized Exchange][sdex]
- kraken ([source](plugins/krakenExchange.go)): [Kraken][kraken]
- binance (_`"ccxt-binance"`_) ([source](plugins/ccxtExchange.go)): Binance via CCXT - only supports priceFeeds and mirroring (buysell, sell, and mirror strategy)
- poloniex (_`"ccxt-poloniex"`_) ([source](plugins/ccxtExchange.go)): Poloniex via CCXT - only supports priceFeeds and mirroring (buysell, sell, and mirror strategy)
- bittrex (_`"ccxt-bittrex"`_) ([source](plugins/ccxtExchange.go)): Bittrex via CCXT - only supports priceFeeds and mirroring (buysell, sell, and mirror strategy)

## Plugins

Kelp can easily be extended because of its _modular plugin based architecture_.
You can create new flavors of the following components: Strategies, PriceFeeds, and Exchanges.

These interfaces make it easy to create plugins:
- Strategy ([source](api/strategy.go)) - API for a strategy
- PriceFeed ([source](api/priceFeed.go)) - API for price of an asset
- Exchange ([source](api/exchange.go)) - API for crypto exchanges

## Directory Structure

The folders are organized to make it easy to find code and streamline development flow.
Each folder is its own package **without any sub-packages**.

    github.com/lightyeario/kelp
    ├── api/            # API interfaces live here (strategy, exchange, price feeds, etc.)
    ├── cmd/            # Cobra commands (trade, exchanges, strategies, etc.)
    ├── examples/       # Sample config files and walkthroughs
    ├── model/          # Low-level structs (dates, orderbook, etc.)
    ├── plugins/        # Implementations of API interfaces (sell strategy, kraken, etc.)
    ├── support/        # Helper functions and utils
    ├── trader/         # Trader bot logic; uses other top-level packages like api, plugins, etc.
    ├── glide.yaml      # Glide dependencies
    ├── main.go         # main function for our kelp binary
    └── ...

## Accounting

You can use [**Stellar-Downloader**][stellar-downloader] to download trade and payment data from your Stellar account as a CSV file.

# Examples

It's easier to learn with examples! Take a look at the walkthrough guides and sample configuration files below.

## Walkthrough Guides

- [Setting up a trading account](examples/walkthroughs/trader/account_setup.md): This guide uses an example token, `COUPON`, to show you how to set up your account before deploying the bot.
- [Market making for a stablecoin](examples/walkthroughs/trader/buysell.md): This guide uses the _buysell_ strategy to provide liquidity for a stablecoin. 
- [ICO sale](examples/walkthroughs/trader/sell.md): This guide uses the `sell` strategy to make a market using sell offers for native tokens in a hypothetical ICO. 
- [Create liquidity for a Stellar-based token](examples/walkthroughs/trader/balanced.md): This guide uses the `balanced` strategy to create liquidty for a token which only trades on the Stellar network. 

## Configuration Files

Reference config files are in the [examples folder](examples/configs/trader). Specifically, the following sample configuration files are included:

- [Sample Sell strategy config file](examples/configs/trader/sample_sell.cfg)
- [Sample BuySell strategy config file](examples/configs/trader/sample_buysell.cfg)
- [Sample Balanced strategy config file](examples/configs/trader/sample_balanced.cfg)
- [Sample Mirror strategy config file](examples/configs/trader/sample_mirror.cfg)

# Contributing

See the [Contribution Guide](CONTRIBUTING.md) and then please [sign the Contributor License Agreement][cla].

# Changelog

See the [Changelog](CHANGELOG.md).

# Questions & Improvements

- Ask questions on the [Stellar StackExchange][stackexchange]; use the `kelp` tag
- [Submit a Bug Report][github-bug-report]
- [Submit a Feature Request][github-feature-request]
- [Raise an issue][github-new-issue] that is not a bug report or a feature request
- [Contribute a PR][github-pulls]

[github-last-commit]: https://github.com/lightyeario/kelp/commit/HEAD
[github-releases]: https://github.com/lightyeario/kelp/releases
[license-apache]: https://opensource.org/licenses/Apache-2.0
[github-issues]: https://github.com/lightyeario/kelp/issues
[github-issues-closed]: https://github.com/lightyeario/kelp/issues?q=is%3Aissue+is%3Aclosed
[github-pulls]: https://github.com/lightyeario/kelp/pulls
[github-pulls-closed]: https://github.com/lightyeario/kelp/pulls?q=is%3Apr+is%3Aclosed
[stellarx]: https://www.stellarx.com
[stablecoin]: https://en.wikipedia.org/wiki/Stablecoin
[stellar q&a]: https://stellarx.zendesk.com/hc/en-us/sections/360001295034-Traders
[scooter video]: https://youtu.be/LStXAG5dwzA
[sdex]: https://www.stellar.org/developers/guides/concepts/exchange.html
[sdex explainer video]: https://www.youtube.com/watch?v=2L8-lrmzeWk
[bash]: https://en.wikipedia.org/wiki/Bash_(Unix_shell)
[golang-download]: https://golang.org/dl/
[golang-setup]: https://golang.org/doc/install#install
[glide-install]: https://github.com/Masterminds/glide#install
[spread]: https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread
[hedge]: https://en.wikipedia.org/wiki/Hedge_(finance)
[cmc]: https://coinmarketcap.com/
[fiat]: https://en.wikipedia.org/wiki/Fiat_money
[currencylayer]: https://currencylayer.com/
[ccxt]: https://github.com/ccxt/ccxt
[ccxt-rest]: https://github.com/franz-see/ccxt-rest
[docker]: https://www.docker.com/
[kraken]: https://www.kraken.com/
[stellar-downloader]: https://github.com/nikhilsaraf/stellar-downloader
[stackexchange]: https://stellar.stackexchange.com/
[cla]: https://goo.gl/forms/lkjJbvkPOO4zZFDp2
[github-bug-report]: https://github.com/lightyeario/kelp/issues/new?template=bug_report.md
[github-feature-request]: https://github.com/lightyeario/kelp/issues/new?template=feature_request.md
[github-new-issue]: https://github.com/lightyeario/kelp/issues/new
