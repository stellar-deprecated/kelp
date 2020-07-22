# TWAP Sale

This guide shows you how to setup the **kelp** bot using the [sell_twap](../../../plugins/sellTwapStrategy.go) strategy. This strategy will sell tokens using the well-known [TWAP metric](https://en.wikipedia.org/wiki/Time-weighted_average_price). We'll configure it to sell a `COUPON` token; `COUPON` can be any token you want and is only used as a sample token for the purpose of this walkthrough guide.

The bot places a single sell order which is refreshed and randomized at every bot update as configured in `trader.cfg`. There is a fixed sell capacity for each `bucket` based on the daily sell limit configuration. Any value that is unsold in a given bucket (surplus) is distributed over the remaining buckets. The config values control the bucket size along with other factors to determine the distribution of the surplus.

## Account Setup

First, go through the [Account Setup guide](account_setup.md) to set up your Stellar accounts and the necessary configuration file, `trader.cfg`. In the `sell_twap` strategy the bot is programmed to sell `ASSET_CODE_A` from the `trader.cfg` file.

## Install Bots

Download the pre-compiled binaries for **kelp** for your platform from the [Github Releases Page](https://github.com/stellar/kelp/releases). If you have downloaded the correct version for your platform you can run it directly.

## SellTwap Strategy Configuration

Use the [sample configuration file for the sell_twap strategy](../../configs/trader/sample_selltwap.cfg) as a template. We will walkthrough the configuration parameters below.

### Price Feeds

SellTwap requires one price feed, `START_ASK_FEED`. This computes the price used when placing the single sell order.

To give `COUPON` a stable price against USD, we're going to set `START_ASK_FEED_TYPE` to `"exchange"` and `START_ASK_FEED_URL` to `"kraken/XXLM/ZUSD"`. This points our bot to Kraken's price for 1 XLM, quoted in USD.

### Tolerances

For the purposes of this walkthrough, we set the `PRICE_TOLERANCE` value to `0.001` which means that a _0.1% change in price will trigger the bot to refresh its orders_. Similarly, we set the `AMOUNT_TOLERANCE` value to `0.001` which means that we need _at least a 0.1% change in amount for the bot to refresh its orders_. If either one of these conditions is met then the bot will refresh its orders. In practice, you should set any value you're comfortable with.

Note that this strategy randomizes the order size at every update, so that should be taken into consideration when setting these tolerance values.

### Buckets

The `PARENT_BUCKET_SIZE_SECONDS` configuration defines the size in seconds of each bucket. There are 86,400 seconds in each day. The number of buckets is determined by dividing 86,400 by `PARENT_BUCKET_SIZE_SECONDS`, so `PARENT_BUCKET_SIZE_SECONDS` needs to perfectly divide 86,400.

In our sample configuration we have set this to 600 seconds, which is equal to 10 minutes. This will give us 144 buckets every day.

### Distributing the Surplus

There are two config params that control how the surplus is distributed, `DISTRIBUTE_SURPLUS_OVER_REMAINING_INTERVALS_PERCENT_CEILING` and `EXPONENTIAL_SMOOTHING_FACTOR`.

`DISTRIBUTE_SURPLUS_OVER_REMAINING_INTERVALS_PERCENT_CEILING` is a decimal value between 0.0 and 1.0, both inclusive. Setting this to 0.0 will discard any surplus and setting it to 1.0 will distribute the surplus over all the remaining buckets. Setting it to 0.50 will distribute it over 50% of the remaining buckets.

`EXPONENTIAL_SMOOTHING_FACTOR` determines how _smoothly_ we should distribute any surplus. This is a decimal value between 0.0 and 1.0, both inclusive. Setting this to 0.0 will sell the entire surplus in the next bucket interval (least smooth). Setting this to 1.0 will distribute the surplus evenly (linearly) over the chosen bucket intervals (most smooth).

### Amounts

The `DAY_OF_WEEK_DAILY_CAP` determines the daily limit of how many tokens of the base asset to sell. This has to be of the form `"volume/daily/sell/base/X/exact"` where `X` is the daily limit in base units to be sold. This configuration is a map of volume filters for each day of the week, which gives you flexibility to have a per-day configuration. A filter for each day must be specified.

The capacity of each bucket is the daily sale amount divided by the number of buckets + the surplus allocated to that bucket.

As mentioned above, the order is randomized and refreshed at each bot update at the interval described in the `trader.cfg` file. The minimum order size for this randomization is controlled by `MIN_CHILD_ORDER_SIZE_PERCENT_OF_PARENT` which is a decimal from 0.0 to 1.0, both inclusive. A value of 0.0 means that there is no minimum order size, whereas a value of 1.0 indicates that the order size should be the capacity of the bucket (i.e. no ramdomization). If the capacity for the bucket is less than the computed minimum amount then the remaining capacity is used as the order size.

## Run Kelp

Assuming your botConfig is called `trader.cfg` and your strategy config is called `selltwap.cfg`, you can run `kelp`  with the following command:

```
kelp trade --botConf ./path/trader.cfg --strategy sell_twap --stratConf ./path/selltwap.cfg
```

# Above and Beyond

You can also play around with the configuration parameters of the [sample configuration file for the sell_twap strategy](../../configs/trader/sample_selltwap.cfg), look at some of the other strategies that are available out-of-the-box or dig into the code and _create your own strategy_.
