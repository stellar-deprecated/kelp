# ICO Sale 

This guide shows you how to setup the **kelp** bot using the [sell](../../../plugins/sellStrategy.go) strategy. We'll configure it to make a market for native tokens we plan to sell in an upcoming hypothetical ICO. This strategy creates sell offers based on a reference price with a pre-specified liquidity depth.  

In our example the tokens are priced against XLM (`native`). XLM is the native currency in the Stellar network and acts as a bridge (i.e. facilitator) in transactions that involve two different assets. Therefore, it has less [counterparty risk](https://www.investopedia.com/terms/c/counterpartyrisk.asp). Most assets have a liquid market against XLM and by pricing your asset against XLM you are opening up your asset to be traded against any other asset on the network via [path payments](https://www.stellar.org/developers/horizon/reference/endpoints/path-finding.html) .

## Account Setup

First, go through the [Account Setup guide](account_setup.md) to set up your Stellar accounts and the necessary configuration file, `trader.cfg`. In the `sell` strategy the bot is programmed to sell `ASSET_CODE_A`. 

## Install Bots

Download the pre-compiled binaries for **kelp** for your platform from the [Github Releases Page](https://github.com/stellar/kelp/releases). If you have downloaded the correct version for your platform you can run it directly.

## Sell Strategy Configuration

Use the [sample configuration file for the sell strategy](../../configs/trader/sample_sell.cfg) as a template. We will walkthrough the configuration parameters below.

### Price Feeds

Sell requires two price feeds, `A` and `B`, and it computes its final trading price by dividing `A` by `B`.

To give `COUPON` a stable price against USD, we're going to connect feed `A` to the **[priceFeed from Kraken](https://kraken.com)**, the popular custodial exchange. We set the `DATA_TYPE_A` to `"exchange"` and `DATA_FEED_A_URL` to `"kraken/XXLM/ZUSD"`. This points our bot to Kraken's price for 1 XLM, quoted in USD. 

By design, we always want to price our stablecoin at Kraken's XLM/USD price. Since that's already what we're getting in feed `A`, and final price is `A`/`B`, we just set price feed `B` to `1.0`. We do this by setting `DATA_TYPE_B` to `"fixed"` and `DATA_FEED_B_URL` to `"1.0"`.

### Tolerances

For the purposes of this walkthrough, we set the `PRICE_TOLERANCE` and `AMOUNT_TOLERANCE` thresholds for our bot to be as small as possible, because we want it to make a lot of trades. We set the value for both fields to `0.001` which means that a _0.1% change in price or amount will trigger the bot to refresh its orders_. In practice, you should set any value you're comfortable with. 

### Trade Amount Base Unit

`AMOUNT_OF_A_BASE` allows you to scale the order sizes set in the next section of the configuration. Trade amounts are specified in **units of the [base asset](https://en.wikipedia.org/wiki/Currency_pair#Base_currency)**.

### Levels

A level defines a [layer](https://en.wikipedia.org/wiki/Layering_(finance)) that sits in the orderbook. Each level has an `AMOUNT` and a `SPREAD` as part of its configuration. The bot creates mirrored orders on both the buy side and the sell side for each level configured.

![level screenshot](https://i.imgur.com/QVjZXGA.png "Levels Screenshot")

`AMOUNT_OF_A_BASE` allows you to scale the order size levels explained below. Trade amounts are specified in **units of the [base asset](https://en.wikipedia.org/wiki/Currency_pair#Base_currency)** (i.e. `ASSET_CODE_A`).

- **AMOUNT**: specifies the order size in multiples of the base unit described above. This `AMOUNT` is multiplied by the `AMOUNT_OF_A_BASE` field to give the final amount. The amount for the quote asset is derived using this value and the computed price at the indicated `SPREAD` level.
- **SPREAD**: represents the distance from the mid price as a percentage specified as a decimal number (0 < spread < 1.00). The [bid/ask spread](https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread) will be 2x what is specified at each level in the config.

## Run Kelp

Assuming your botConfig is called `trader.cfg` and your strategy config is called `sell.cfg`, you can run `kelp`  with the following command:

```
kelp trade --botConf ./path/trader.cfg --strategy sell --stratConf ./path/sell.cfg
```

# Above and Beyond

You can also play around with the configuration parameters of the [sample configuration file for the sell strategy](../../configs/trader/sample_sell.cfg), look at some of the other strategies that are available out-of-the-box or dig into the code and _create your own strategy_.
