# Kelp

[![GitHub last commit](https://img.shields.io/github/last-commit/lightyeario/kelp.svg?style=for-the-badge)][github-last-commit]
[![Github All Releases](https://img.shields.io/github/downloads/lightyeario/kelp/total.svg?style=for-the-badge)][github-releases]
[![license](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge&longCache=true)][license-apache]

[![GitHub issues](https://img.shields.io/github/issues/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-issues]
[![GitHub closed issues](https://img.shields.io/github/issues-closed/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-issues-closed]
[![GitHub pull requests](https://img.shields.io/github/issues-pr/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-pulls]
[![GitHub closed pull requests](https://img.shields.io/github/issues-pr-closed/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-pulls-closed]

Kelp is a free and open-source trading bot for the [Stellar universal marketplace][stellarx].

Kelp includes several configurable trading strategies, and its modular design allows you to customize your algorithms, exchange integrations, and assets. You can define your own parameters or use the existing repository to quickly implement a trading bot. With Kelp, you could be up and trading on Stellar in a matter of minutes.

Kelp is pre-configured to:

- Make spreads and make markets
- Create liquidity and facilitate price-discovery for ICOs
- Price and trade custom [stablecoins][stablecoin]
- Mimic order books from other exchanges

To learn more about the Stellar protocol check out [this community video][sdex explainer video]. You can also search [Stellar's Q&A][stellar q&a].

# Table of Contents

   * [Be Smart and Go Slow](#be-smart-and-go-slow)
   * [Getting Started](#getting-started)
      * [Compile from Source](#compile-from-source)
      * [Running the Trader Bot](#running-the-trader-bot)
   * [Components](#components)
      * [Strategies](#strategies)
      * [Price Feeds](#price-feeds)
      * [Configuration Files](#configuration-files)
      * [Exchanges](#exchanges)
      * [Extensions](#extensions)
   * [Examples](#examples)
      * [Walkthrough Guides](#walkthrough-guides)
      * [Configuration Files](#configuration-files-1)
   * [Questions &amp; Improvements](#questions--improvements)

# Be Smart and Go Slow

_Whenever you trade on Stellar, you are trading with volatile assets, in volatile markets, and you risk losing money. Period. Use Kelp at your own risk. There is no guarantee you'll make a profit from using our bots or strategies. In fact, if you set bad parameters or market conditions change, Kelp might help you **lose** money very fast. So be smart and go slow._

# Getting Started

To get started with Kelp, either download the pre-compiled binary for your platform from the [Github Releases Page][github-releases] or [compile Kelp from source](#compile-from-source).

There is **one** binary associated with this project: [`trader`](trader). Once the binary is downloaded, run the bot by following the instructions in [Running the Trader Bot](#running-the-trader-bot).

## Compile from Source

To compile Kelp from source:

1. [Download][golang-download] and [setup][golang-setup] Golang.
2. [Install Glide][glide-install] for dependency management.
3. Clone the repo into `$GOPATH/src/github.com/lightyeario/kelp`: `git clone git@github.com:lightyeario/kelp.git`
4. Change to the kelp directory and install the dependencies: `glide install`
5. Build the binaries: `go install github.com/lightyeario/kelp/...`
6. Confirm _one new binary_ in `$GOPATH/bin`: `trader`.

## Running the Trader Bot

The Trader Bot places orders on the [Stellar marketplace][stellarx] based on the selected strategy. Configuration files specify the Stellar account and strategy details.

`trader` has three required parameters which are:

- **botConf**: `.cfg` file with the account details as defined [here](trader/config.go).
- **stratType**: the strategy you want to run (`sell`, `buysell`, `balanced`, `mirror`, `delete`).
- **stratConf**: `.cfg` file specific to your chosen strategy, find the [strategies here](trader/strategy).

Example:

`trader --botConf traderConf.cfg --stratType buysell --stratConf buysell.cfg`

# Components

Kelp includes an assortment of strategies, price feeds, and extensions you can build around. Kelp also enables you to create your own trading strategies.

## Strategies

Strategies are at the core of the **trader bot**. Without them it's just lazy, capable of nothing, thinking of nothing, doing nothing, like our friend [scooter][scooter video] here. The strategies give your bot purpose. Each approaches the market in a different way and is designed to achieve a particular goal.

The following [strategies](trader/strategy) are available **out of the box** with Kelp:

- [sell](trader/strategy/sell.go):

    - **What:** creates sell offers based on a reference price with a pre-specified liquidity depth
    - **Why:** To sell tokens at a fixed price or at a price that changes based on an external reference price
    - **Who:** An issuer could use Sell to distribute tokens from an ICO pre-sale
    - **Complexity**: Beginner

- [buysell](trader/strategy/buysell.go):

    - **What:** creates buy and sell offers based on a specific reference price and a pre-specified liquidity depth while maintaining a [spread][spread].
    - **Why:** To make the market for tokens based on a fixed or external reference price.
    - **Who:** Anyone who wants to create liquidity for a stablecoin or [fiat][fiat] token
    - **Complexity:** Beginner

- [balanced](trader/strategy/balanced.go):
    - **What:** dynamically prices two tokens based on their relative demand. For example, if more traders buy token A _from_ the bot (the traders are therefore selling token B), the bot will automatically raise the price for token A and drop the price for token B.
    - **Why:** To let the market surface the _true price_ for one token in terms of another.
    - **Who:** Market makers and traders for tokens that trade only on Stellar 
    - **Complexity:** Intermediate

- [mirror](trader/strategy/mirror.go):

    - **What:** mirrors an orderbook from another exchange by placing the same orders on Stellar after including a [spread][spread]. _Note: covering your trades on the backing exchange is not currently supported out-of-the-box_.
    - **Why:** To [hedge][hedge] your position on another exchange whenever a trade is executed to reduce inventory risk while keeping a spread
    - **Who:** Anyone who wants to reduce inventory risk and also has the capacity to take on a higher operational overhead in maintaining the bot system.
    - **Complexity:** Advanced

- [delete](trader/strategy/delete.go):

    - **What:** deletes your offers from both sides of the specified orderbook. _Note: does not need a strategy-specific config file_.
    - **Why:** To kill the offers placed by the bot. _This is not a trading strategy but is used for operational purposes only_.
    - **Who:** Anyone managing the operations of the bot who wants to stop all activity by the bot.
    - **Complexity:** Beginner

## Price Feeds

Price Feeds fetch the price of an asset from an external source. The following [price feeds](support/priceFeed) are available **out of the box** with Kelp:

- [coinmarketcap](support/priceFeed/cmcFeed.go): fetches the price of tokens from [CoinMarketCap][cmc]
- [fiat](support/priceFeed/fiatFeed.go): fetches the price of a [fiat][fiat] currency from the [CurrencyLayer API][currencylayer]
- [exchange](support/priceFeed/exchange.go): fetches the price from an exchange
- [fixed](support/priceFeed/fixedFeed.go): sets the price to a constant

## Configuration Files

Each strategy you implement needs a configuration file. The format of the configuration file is specific to the selected strategy. You can use these files to customize parameters for your chosen strategy. For more details, check out the [examples section](#configuration-files) of the readme.

## Exchanges

Exchange integrations provide data to trading strategies and allow you to [hedge][hedge] your positions on different exchanges. The following [exchange integrations](support/exchange) are available **out of the box** with Kelp:

- [sdex](support/utils/txbutler.go): The [Stellar Decentralized Exchange][sdex]
- [kraken](support/exchange/kraken): [Kraken][kraken]

## Extensions

Kelp can easily be extended because of its _modular plugin based architecture_.
You can create new flavors of the following components: Strategies, PriceFeeds, and Exchanges.

These interfaces make it easy to create extensions:
- [Strategy](api/strategy.go) - API for a strategy used by the `trader` bot
- [SideStrategy](api/sideStrategy.go) - API for a strategy that is applied to a single side of the orderbook. Can be used in conjunction with [compose.go](trader/strategy/compose.go) to build a Strategy using principles of [composition][composition].
- [PriceFeed](support/priceFeed/pricefeed.go#L13) - API for price of an asset
- [Exchange](api/exchange.go#L30) - API for crypto exchanges

# Examples

It's easier to learn with examples! Take a look through the walkthrough guides and sample configuration files below.

## Walkthrough Guides

- [Setting up a trading account](examples/walkthroughs/trader/account_setup.md): This guide uses an example token, `COUPON`, to show you how to set up your account before deploying the bot.
- [Market making for a stablecoin](examples/walkthroughs/trader/buysell.md): This guide shows you how to use the bot with the _buysell_ strategy.

## Configuration Files

Reference config files are in the [examples folder](examples/). Specifically, the following sample configuration files are included:

- [Sample Sell strategy config file](examples/configs/trader/sample_sell.cfg)
- [Sample BuySell strategy config file](examples/configs/trader/sample_buysell.cfg)
- [Sample Balanced strategy config file](examples/configs/trader/sample_balanced.cfg)
- [Sample Mirror strategy config file](examples/configs/trader/sample_mirror.cfg)

# Questions & Improvements

- Ask questions on the [Stellar StackExchange][stackexchange]; use the `kelp` tag
- [Raise an issue][github-issues]
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
[stellar q&a]: https://www.stellar.org
[scooter video]: https://youtu.be/LStXAG5dwzA
[sdex]: https://www.stellar.org/developers/guides/concepts/exchange.html
[sdex explainer video]: https://www.youtube.com/watch?v=2L8-lrmzeWk
[golang-download]: https://golang.org/dl/
[golang-setup]: https://golang.org/doc/install#install
[glide-install]: https://github.com/Masterminds/glide#install
[spread]: https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread
[hedge]: https://en.wikipedia.org/wiki/Hedge_(finance)
[cmc]: https://coinmarketcap.com/
[fiat]: https://en.wikipedia.org/wiki/Fiat_money
[currencylayer]: https://currencylayer.com/
[kraken]: https://www.kraken.com/
[stackexchange]: https://stellar.stackexchange.com/
[composition]: https://en.wikipedia.org/wiki/Object_composition
