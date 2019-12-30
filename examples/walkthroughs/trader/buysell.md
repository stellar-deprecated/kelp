# How To Make a Market for a Stablecoin

This guide shows you how to setup the **kelp** bot using the [buysell](../../../plugins/buysellStrategy.go) strategy. We'll configure it to make a market for a stablecoin against the XLM asset.

This strategy buys low and sells high with a pre-defined [spread](https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread) and [priceFeed](../../../api/priceFeed.go).

## Account Setup

First, go through the [Account Setup guide](account_setup.md) to set up your Stellar accounts and the necessary configuration file, `trader.cfg`. This introduces the `COUPON` token that we will use through the rest of this walkthrough.

## Install Bots

Download the pre-compiled binaries for **kelp** for your platform from the [Github Releases Page](https://github.com/stellar/kelp/releases). If you have downloaded the correct version for your platform you can run it directly.

## BuySell Strategy Configuration

Use the [sample configuration file for the buysell strategy](../../configs/trader/sample_buysell.cfg) as a template. We will walkthrough the configuration parameters below.

### Price Feeds

BuySell requires two price feeds and computes its final trading price by combining these price feeds as `price(A)`/`price(B)`.

To give `COUPON` a stable price against USD, we're going to connect feed `A` to the **[priceFeed from Kraken](https://kraken.com)**, the popular custodial exchange. We set the `DATA_TYPE_A` to `"exchange"` and `DATA_FEED_A_URL` to `"kraken/XXLM/ZUSD"`. This points our bot to Kraken's price for 1 XLM, quoted in USD. 

We always want to price our stablecoin at Kraken's XLM/USD price. Since that's already what we're getting in feed `A`, we want to set price feed `B` to `1.0`. We do this by setting `DATA_TYPE_B` to `"fixed"` and `DATA_FEED_B_URL` to `"1.0"`.

### Tolerances

For the purposes of this walkthrough, we set the `PRICE_TOLERANCE` and `AMOUNT_TOLERANCE` thresholds for our bot to be as small as possible because we want it to make a lot of trades. We set the value for both fields to `0.001` which means that a _0.1% change in price or amount will trigger the bot to refresh its orders_. In practice, you should set any value you're comfortable with. 

### Levels

A level defines a [layer](https://en.wikipedia.org/wiki/Layering_(finance)) that sits in the orderbook. Each level has an `AMOUNT` and a `SPREAD` as part of its configuration. The bot creates mirrored orders on both the buy side and the sell side for each level configured.

![level screenshot](https://i.imgur.com/QVjZXGA.png "Levels Screenshot")

`AMOUNT_OF_A_BASE` allows you to scale the order size levels explained below. Trade amounts are specified in **units of the [base asset](https://en.wikipedia.org/wiki/Currency_pair#Base_currency)** (i.e. `ASSET_CODE_A`).

- **SPREAD**: represents the distance from the mid price as a percentage specified as a decimal number (0 < spread < 1.00). The [bid/ask spread](https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread) will be 2x what is specified at each level in the config.
- **AMOUNT**: specifies the order size in multiples of the base unit described above. This `AMOUNT` is multiplied by the `AMOUNT_OF_A_BASE` field to give the final amount. The amount for the quote asset is derived using this value and the computed price at this level. 

## Run Kelp

Assuming your botConfig is called `trader.cfg` and your strategy config is called `buysell.cfg`, you can run `kelp` with the following command:
```
kelp trade --botConf ./path/trader.cfg --strategy buysell --stratConf ./path/buysell.cfg
```

# Above and Beyond

You can also play around with the configuration parameters of the [sample configuration file for the buysell strategy](../../configs/trader/sample_buysell.cfg), look at some of the other strategies that are available out-of-the-box or dig into the code and _create your own strategy_.
