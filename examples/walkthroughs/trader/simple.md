# Setting up the Trader Bot with the Simple Strategy

This guide walks through the set up of a `trader` bot using the [simple](../../../trader/strategy/simple.go) strategy.

This strategy buys low and sells high with a pre-defined [spread](https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread) and [priceFeed](../../../support/priceFeed/pricefeed.go).

## Account Setup

You should first go through the [Account Setup Walkthrough](account_setup.md) to set up your Stellar accounts and botConfig. This introduces the `COUPON` token that we will use through the rest of this walkthrough.

## Install Bots

Download the pre-compiled binaries for the **trader bot** for your platform from the [Github Releases Page](https://github.com/lightyeario/kelp/releases).
If you have downloaded the correct version for your platform you can run it directly.

## Simple Strategy Configuration

You can use the [sample configuration file for the simple strategy](../../configs/trader/sample_simple.cfg) as a template.
The sample config file has the data for this walkthrough pre-filled, we describe how we got to these configuration parameters in the sub-sections below.

### Price Feeds

We are going to use the **priceFeed from the Kraken Exchange** to price our `COUPON` tokens to be stable with the price of the US Dollar.
In order to achieve this, we set the `DATA_TYPE_A` to `"exchange"` and `DATA_FEED_A_URL` to `"kraken/XXLM/ZUSD"`. This represents the kraken exchange providing a price of 1 XLM **quoted in** the USD Currency. i.e. _the number represents how many US Dollars I can get in exchange for 1 XLM_.

We have a second price feed parameter in the config that we **are required to set**. The final price is computed by dividing the price received from the first feed `A` by the price received from the second feed `B`. _Since the Kraken feed is already providing us with a ratio of XLM/USD we can set the feed `B` to be a fixed value of `1.0`_. This can be achieved by setting `DATA_TYPE_B` to `"fixed"` and `DATA_FEED_B_URL` to `"1.0"`.

### Tolerances

We want to set the `PRICE_TOLERANCE` and `AMOUNT_TOLERANCE` levels to be as little as possible in our example because we want our bot to place more orders for the purposes of this walkthrough. For that reason we set the value for both fields to `0.001` which means that a _0.1% change in price or amount will trigger the bot to refresh the order placed at that level_.

### Scaling Amounts

`AMOUNT_OF_A_BASE` can be used to scale the amount levels set in the next section of the configuration. This is mostly just a convenience to increase/decrease the size of the orders placed by the bot.

Note: amounts are specified in **units of the base asset**.

### Levels

Levels are mirrored on the buy and sell side. Each level has a `SPREAD` and an `AMOUNT`.

- **SPREAD**: represents the distance from center price as a percentage specified as a decimal number (0 < spread < 1.00). The bid/ask spread will be 2x what is specified at each level in the config.
- **AMOUNT**: specifies the order size as the number of units of the base asset to be placed at this level. This amount is multiplied by the `AMOUNT_OF_A_BASE` field to give the final amount for this level. The amount for the quote asset is derived using this value and the computed price at that spread level.

## Run the Trader Bot

Assuming your botConfig is called `traderConf.cfg` and your strategy config is called `simple.cfg`, you can run the `trader` bot with the following command:
```
trader --botConf traderConf.cfg --stratType simple --stratConf simple.cfg
```

If you want to use a different trading strategy, you can change the `stratType` and provide the relevant config file for your chosen strategy.

## Next Steps

After completing this guide successfully, you should have some experience running the `trader` bot using the [**simple strategy**](../../../trader/strategy/simple.go).

You can play around with the configuration parameters of the [sample configuration file for the simple strategy](../../configs/trader/sample_simple.cfg), look at some of the other strategies that are available out-of-the-box or [dig into the code](../../../trader/strategy) and _create your own strategy_.
