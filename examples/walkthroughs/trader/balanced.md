# Create Liquidity For a Stellar-based Token

This guide shows you how to setup the **kelp** bot using the [balanced](../../../plugins/balancedStrategy.go) strategy. We'll configure it to create liquidity for a `COUPON` token which only trades on the Stellar network. 

The bot dynamically prices two tokens based on their relative demand. It operates with the understanding that both of the assets it holds are equal in value. When someone buys the [base asset](https://en.wikipedia.org/wiki/Currency_pair#Base_currency) from the bot, the bot will have less units of the base asset and more units of the counter asset. It will assume that the base asset is now more valuable than the counter asset and will raise the price of the base asset relative to the counter asset. 

In our scenario, our bot holds `COUPON` and XLM. If more traders buy `COUPON` from the bot (and are selling XLM), the bot will automatically raise the price for `COUPON` and lower it for XLM. 

## Account Setup

First, go through the [Account Setup guide](account_setup.md) to set up your Stellar accounts and the necessary configuration file, `trader.cfg`.

**The account needs to be funded with equivalent values of both assets. For example, assuming 1 `XLM` = 3 `COUPON` then the bot should be funded with 10,000 `XLM` and 30,000 `COUPON`. It is important to use a reliable and trusted source when picking the initial value for both assets.**

## Install Bots

Download the pre-compiled binaries for **kelp** for your platform from the [Github Releases Page](https://github.com/stellar/kelp/releases). If you have downloaded the correct version for your platform you can run it directly.

## Balanced Strategy Configuration

Use the [sample configuration file for the balanced strategy](../../configs/trader/sample_balanced.cfg) as a template. We will walk through the configuration parameters below.

### Tolerances

For the purposes of this walkthrough, we set the `PRICE_TOLERANCE` and `AMOUNT_TOLERANCE` thresholds for our bot to be as small as possible because we want it to make a lot of trades. We set the value for both fields to `0.001` which means that a _0.1% change in price or amount will trigger the bot to refresh its orders_. In practice, you should set any value you're comfortable with. 

### Spread

- **`SPREAD`**: refers to the [bid/ask spread](https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread) as a percentage represented as a decimal number (`0.0` < spread < `1.00`).
- **`MIN_AMOUNT_SPREAD`** & **`MAX_AMOUNT_SPREAD`**: reduce the order size by the indicated percentage (specified as a decimal). If someone buys and subsequently sells the full order amount you will end up with a profit equalling this percentage multiplied by the full order amount. This effectively makes the spread. 

### Levels 

A level defines a [layer](https://en.wikipedia.org/wiki/Layering_(finance)) that sits in the orderbook. 

![level screenshot](https://i.imgur.com/QVjZXGA.png "Levels Screenshot")

- **`MAX_LEVELS`**: defines the depth of your order book by indicating the maximum number of levels on the buy and sell side
- **`LEVEL_DENSITY`**: a value between `0.0` and `1.0` used as a probability of adding orders at a given price level. Setting this to `1.0` will make the depth chart look more like steps. Doing so will make it obvious that your orders are created by bots so feel free to play around with this value.
- **`ENSURE_FIRST_N_LEVELS`**: ensures the first N levels always exist on either side of the order book. 

If your `LEVEL_DENSITY` is < `1.0` the bot will accumulate the amounts that it would have otherwise placed into a variable called `amountCarryover`. As the bot places more offers, it first decides whether it should place an order at the given level using randomness controlled by the `CARRYOVER_INCLUSION_PROBABILITY` parameter.

The `amountCarryoverSpread` determines how much of the `amountCarryover` should be consumed. Randomness is used when picking the `amountCarryoverSpread` bounded by the `MIN_AMOUNT_CARRYOVER_SPREAD` and `MAX_AMOUNT_CARRYOVER_SPREAD` parameters. As you increase these values, the depth chart will look taller between levels accordingly. Setting these parameters will provide randomness to the offers while still keeping a deep orderbook. _Before setting these values you should test them first._ 
 
- **`CARRYOVER_INCLUSION_PROBABILITY`**: a decimal number between `0.0` and `1.0` as defined above 
- **`MIN_AMOUNT_CARRYOVER_SPREAD` and `MAX_AMOUNT_CARRYOVER_SPREAD`**: a decimal number between `0.0` and `1.0` as defined above

### Virtual Balance 

_These parameters are dangerous. It is recommended to set these parameters to `0.0`_.

Setting a virtual balance for any of the two assets fools the bot into thinking that it has a bigger balance in its account for that asset. Doing so will result in the bot setting the levels with a smoother pricing curve. 

If you set these values to `0.0` and the bot happens to sell all but the last unit of the asset, the last asset will be valued at _infinity_ by the bot and it will be almost impossible for the bot to sell the last unit. However, if you set this value to be > `0.0` the bot will eventually run out of the asset that has a virtual balance set and the bot will get stuck. The behavior of the bot in this state is _undefined_. 

The virtual balance combined with the actual balance the bot has in its account will be used to compute the _total balance_ for that asset. The _total balance_ for a particular asset is used when computing the relative the price between the assets using the ratio of their balances. 

- **`VIRTUAL_BALANCE_BASE`**: a decimal value 
- **`VIRTUAL_BALANCE_QUOTE`**: a decimal value 

## Run Kelp

Assuming your botConfig is called `trader.cfg` and your strategy config is called `balanced.cfg`, you can run `kelp` with the following command:

```
kelp trade --botConf ./path/trader.cfg --strategy balanced --stratConf ./path/balanced.cfg
```

# Above and Beyond

You can also play around with the configuration parameters of the [sample configuration file for the balanced strategy](../../configs/trader/sample_balanced.cfg), look at some of the other strategies that are available out-of-the-box or dig into the code and _create your own strategy_.
