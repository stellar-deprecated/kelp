# Kelp

[![GitHub last commit](https://img.shields.io/github/last-commit/lightyeario/kelp.svg?style=for-the-badge)][github-last-commit] [![Github All Releases](https://img.shields.io/github/downloads/lightyeario/kelp/total.svg?style=for-the-badge)][github-releases] [![license](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge&longCache=true)][license-apache]

[![GitHub issues](https://img.shields.io/github/issues/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-issues] [![GitHub closed issues](https://img.shields.io/github/issues-closed/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-issues-closed]

[![GitHub pull requests](https://img.shields.io/github/issues-pr/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-pulls] [![GitHub closed pull requests](https://img.shields.io/github/issues-pr-closed/lightyeario/kelp.svg?style=flat-square&longCache=true)][github-pulls-closed]

Kelp is an open source trading bot for the [Stellar Network][stellar.org].

Kelp was built with the intention to plug in any custom strategy, exchange integration, or even any asset. This can be seen in the **modular design** of the project. The idea is that anyone can bring their own strategies or use the existing repository of strategies and quickly get up and running with a trading bot on the [Stellar Decentralized Exchange][sdex].

# Stellar Decentralized Exchange (SDEX)

Here's an explainer video created by a community member:
[![SDEX Explainer Video][sdex explainer image]][sdex explainer video]

# Getting Started

You can download pre-compiled binaries for your platform from the [Github Releases Page][github-releases] or [Compile from Source](#compile-from-source). There is **one** binary associated with this project: [`trader`](trader).

Once you have the binary, you can run the bot by following the instuctions in [Running the Trader Bot](#running-the-trader-bot).

## Compile from Source

Here are the steps to compile from source:

1. [Download][golang-download] and [Setup][golang-setup] Golang.
2. We use Glide for dependency management, [install from here][glide-install].
3. Clone the repo into `$GOPATH/src/github.com/lightyeario/kelp`: `git clone git@github.com:lightyeario/kelp.git`
4. Build the binaries: `go install github.com/lightyeario/kelp/...`
5. You will now have _one new binary_ in `$GOPATH/bin`: `trader`.

## Running the Trader Bot

The Trader Bot places orders on the [Stellar Decentralized Exchange][sdex] based on the selected strategy. It uses configuration files to take in the Stellar Account and Strategy Configurations.

These are the required parameters to `trader`:

- **botConf**: `.cfg` file with the account details as defined [here](trader/config.go).
- **stratType**: type of strategy to be run (`simple`, `sell`, `mirror`, `autonomous`).
- **stratConf**: strategy-dependent `.cfg` config file defined per strategy, you can find the [strategies here](trader/strategy).

Example: `trader --botConf traderConf.cfg --stratType simple --stratConf simple.cfg`

# What You Get Out of the Box

## Strategies

Strategies are at the core of the **trader bot**. Each strategy can be used for different goals and approaches the market in the unique way.

_Disclaimer: If you use these strategies please do so at your own risk. You are trading with assets that have monetary value and you risk losing money when you trade. We **do not claim** that you will make a profit by using any of these strategies._

Here are some of the [strategies](trader/strategy) that are available **out of the box** with Kelp:

- [sell](trader/strategy/sell.go): sell an asset at a specific price
    - **Complexity**: Beginner
    - **Explanation**: Creates sell offers for a digital asset based on a specific reference price with the pre-specified liquidity depth
    - **When to Use**: You want to sell your tokens at a fixed price or a price that changes based on an external reference price.
    - **Example**: Distributing your tokens in an ICO pre-sale by creating sell offers on the DEX.
- [simple](trader/strategy/simple.go): buy and sell an asset at a specific price with a [spread][spread]
    - **Complexity**: Beginner
    - **Explanation**: This is a market making strategy that creates buy and sell offers for a digital asset based on a specific reference price with the pre-specified liquidity depth
    - **When to Use**: You want to make the market for your tokens based on a fixed or external reference price.
    - **Example**: Creating liquidity for your stable coin or fiat token on the DEX.
- [autonomous](trader/strategy/autonomous.go): dynamic pricing of [ICO][ico] tokens based on demand
    - **Complexity**: Intermediate
    - **Explanation**: This strategy creates buy and sell offers based on the ratio of the asset balances held by the bot's account. This bot mimics the laws of supply and demand; if more people buy asset A **from** the bot (i.e. sell asset B to the bot) then the bot will automatically raise the price for asset A (and drop the price for asset B) while maintaining the pre-defined spread.
    - **When to Use**: You have a token that you want the market to price for you, or you are operating in a liquid market where the price is fairly stable.
    - **Example**: Making the market for your ICO token that is available only on the DEX or trading such a token.
- [mirror](trader/strategy/mirror.go): mirror liquidity from one exchange to another with a small [spread][spread].
    - **Complexity**: Advanced
    - **Explanation**: Mirrors the orderbook on another exchange by placing the same orders on SDEX. _Note: covering your trades on the backing exchange is not currently supported out-of-the-box_.
    - **When to Use**: You want to [hedge][hedge] your position whenever someone trades with you to reduce inventory risk while keeping a spread.
    - **Example**: You are a risk-taking individual or firm who is willing to be the market maker for a market for the possibility of making a profit on the spread of each trade.

## Price Feeds

Price Feeds allow you to fetch the price of an asset from an external source. Here are some of the [price feeds](support/priceFeed) that are available **out of the box** with Kelp:

- [coinmarketcap](support/priceFeed/cmcFeed.go): fetches the price from [CoinMarketCap][cmc]
- [fiat](support/priceFeed/fiatFeed.go): fetches the price of a [fiat][fiat] currency from the [CurrencyLayer API][currencylayer]
- [exchange](support/priceFeed/exchange.go): fetches the price from an exchange
- [fixed](support/priceFeed/fixedFeed.go): sets the price to a constant

## Exchanges

Here are some of the [exchange integrations](support/exchange) available **out of the box** with Kelp:

- [sdex](support/txbutler.go): [Stellar Decentralized Exchange][sdex]
- [kraken](support/exchange/kraken): [Kraken][kraken]

## Extensions

All the above parts (strategies, price feeds, exchanges) can be easily extended to create _new plugins_.

An example of the tools available to make it easier to build strategies is [compose.go](trader/strategy/compose.go) which [composes][composition] two instances of the [Side Strategy API](trader/strategy/sideStrategy/sideStrategy.go) to give you an instance of a [Strategy](trader/strategy/strategy.go).

# Examples

### Configs

You can find [sample config files for the trader bot here](examples/configs/trader).

### Walkthrough Guides

We have prepared some [walkthough guides for the trader bot](examples/walkthroughs/trader) that may be useful in getting started.

# Questions

You can ask any questions on the [Stellar StackExchange][stackexchange] with the `kelp` tag or [raise an issue][github-issues] in this repo.

You can chat on the `kelp` channel in the [Stellar Public Slack][stellar-slack]. If you don't have an account you can [create an account here][stellar-slack-new].

[github-last-commit]: https://github.com/lightyeario/kelp/commit/HEAD
[github-releases]: https://github.com/lightyeario/kelp/releases
[license-apache]: https://opensource.org/licenses/Apache-2.0
[github-issues]: https://github.com/lightyeario/kelp/issues
[github-issues-closed]: https://github.com/lightyeario/kelp/issues?q=is%3Aissue+is%3Aclosed
[github-pulls]: https://github.com/lightyeario/kelp/pulls
[github-pulls-closed]: https://github.com/lightyeario/kelp/pulls?q=is%3Apr+is%3Aclosed
[stellar.org]: https://www.stellar.org/
[sdex]: https://www.stellar.org/developers/guides/concepts/exchange.html
[sdex explainer image]: http://img.youtube.com/vi/2L8-lrmzeWk/0.jpg
[sdex explainer video]: https://www.youtube.com/watch?v=2L8-lrmzeWk
[golang-download]: https://golang.org/dl/
[golang-setup]: https://golang.org/doc/install#install
[glide-install]: https://github.com/Masterminds/glide#install
[spread]: https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread
[ico]: https://en.wikipedia.org/wiki/Initial_coin_offering
[hedge]: https://en.wikipedia.org/wiki/Hedge_(finance)
[cmc]: https://coinmarketcap.com/
[fiat]: https://en.wikipedia.org/wiki/Fiat_money
[currencylayer]: https://currencylayer.com/
[kraken]: https://www.kraken.com/
[stackexchange]: https://stellar.stackexchange.com/
[composition]: https://en.wikipedia.org/wiki/Object_composition
[stellar-slack]: https://stellar-public.slack.com/messages
[stellar-slack-new]: https://slack.stellar.org/
