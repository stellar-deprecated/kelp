# Create Targeted Liquidity Within A Bounded Price Range

This guide shows you how to setup the **kelp** bot using the [pendulum](../../../plugins/pendulumStrategy.go) strategy. We'll configure it to create liquidity for a `COUPON` token. `COUPON` can be any token you want and is only used as a sample token for the purpose of this walkthrough guide.

The bot dynamically prices two tokens based on their relative demand. When someone _buys_ the [base asset](https://en.wikipedia.org/wiki/Currency_pair#Base_currency) from the bot (i.e. the bot's ask order is executed resulting in the bot selling the base asset), the bot will _increase_ the price of the base asset relative to the counter asset. Conversely, when someone _sells_ the [base asset](https://en.wikipedia.org/wiki/Currency_pair#Base_currency) to the bot (i.e. the bot's bid order is executed resulting in the bot buying the base asset), the bot will _decrease_ the price of the base asset relative to the counter asset.

This strategy operates like a pendulum in that if a trader hits the ask then the bot moves the price to the right side (higher), if a trader hits the bid then the bot moves the price to the left side (lower).

## Account Setup

First, go through the [Account Setup guide](account_setup.md) to set up your Stellar accounts and the necessary configuration file, `trader.cfg`.

## Install Bots

Download the pre-compiled binaries for **kelp** for your platform from the [Github Releases Page](https://github.com/stellar/kelp/releases). If you have downloaded the correct version for your platform you can run it directly.

## Pendulum Strategy Configuration

Use the [sample configuration file for the pendulum strategy](../../configs/trader/sample_pendulum.cfg) as a template. We will walk through the configuration parameters below.

### Tolerances

For the purposes of this walkthrough, we set the `PRICE_TOLERANCE` value to `0.001` which means that a _0.1% change in price will trigger the bot to refresh its orders_. Similarly, we set the `AMOUNT_TOLERANCE` value to `1.0` which means that we need _at least a 100% change in amount will trigger the bot to refresh its orders_, i.e. an order needs to be fully consumed to trigger a replacement of that order. If either one of these conditions is met then the bot will refresh its orders.

In practice, you should set any value you're comfortable with.

### Amounts

- **`AMOUNT_BASE_BUY`**: refers to the order size denominated in units of the base asset to be placed for the bids, represented as a decimal number.
- **`AMOUNT_BASE_SELL`**: refers to the order size denominated in units of the base asset to be placed for the asks, represented as a decimal number.

### Spread

- **`SPREAD`**: refers to the [bid/ask spread](https://en.wikipedia.org/wiki/Bid%E2%80%93ask_spread) as a percentage represented as a decimal number (`0.0` < spread < `1.00`).

**Note 1**: the resting bid and ask orders will have a larger spread than what is specified in the config.
The reason is that the bids and asks adjust automatically by moving up/down when orders are taken as described above.
If an ask is taken then all bid and ask orders move up. If a bid is taken then all bid and ask orders move down.
This `SPREAD` config value is the effective spread percent you will receive after adjusting for the automatic movement of orders, assuming that the bot has time to move the orders before the next trade happens. If the bot does not have time to move the orders then the bot will receive a larger spread than what is specified in the config.

**Note 2**: this spread value percent should be greater than or equal to **2 x fee** on the exchange you are trading on.
The intuition behind this is that in order to complete a roundtrip (buy followed by sell, or sell followed by buy), you will make two trades which will cost you **2 x fee** as a percentage of your order size.
By setting a spread value greater than or equal to **2 x fee** you are accounting for the fees as a cost of your trading activities.

### Levels

A level defines a [layer](https://en.wikipedia.org/wiki/Layering_(finance)) that sits in the orderbook. The bot will create mirrored orders on both the buy side and the sell side for each level configured.

- **`MAX_LEVELS`**: refers to the number of levels that you want on either side of the mid-price. 

![level screenshot](https://i.imgur.com/QVjZXGA.png "Levels Screenshot")

### Price Limits

It is important to set price limits to control for changing market conditions. **It is highly recommended to set all three of these values. It is extremely dangerous to not set them.**

- **`SEED_LAST_TRADE_PRICE`**: (required) price with which to start off as the last trade price (i.e. initial center price). A good value for this is the current mid-price of the market you are trading on, but it is not always the best choice.
- **`MAX_PRICE`**: maximum price to offer, without this setting you could end up at a price where your algorithm is no longer effective
- **`MIN_PRICE`**: minimum price to offer, without this setting you could end up at a price where your algorithm is no longer effective

**Note 1**: You are required to set the `SEED_LAST_TRADE_PRICE` otherwise the algorithm will not work. **it is highly recommend to always update the value of `SEED_LAST_TRADE_PRICE` in the configuration before starting a new run of the bot so you can ensure it is line with the current market price.**

**Note 2**: It is recommended to set the `MAX_PRICE` and the `MIN_PRICE` to define the bounds in which your algorithm will work as expected. If market conditions change to the point where the price of the market goes outside this range then your configuration is no longer valid and it is better for your bot to pause trading.
This can be caused by a spike in the relative value of one asset compared to the other, which is not conducive to this trading strategy and you should re-evaluate and update your configuration in this situation, or consider stopping your trading activities on this market altogether.

### Minimum Amount Limits

You may want to ensure that your account has a minimum balance of a given asset so you do not risk running out of any one asset. These settings help you configure that.

- **`MIN_BASE`**: decimal value representing the minimum amount of base asset balance to maintain after which the strategy won't place any more orders
- **`MIN_QUOTE`**: decimal value representing the minimum amount of quote asset balance to maintain after which the strategy won't place any more orders

### Historical Trades

This trading strategy adjusts the offered price based on the last price it received for a trade. In order to do this it needs to fetch trades from the exchange. In order to do this the bot will need to know from which point to start fetching trades (_cursor_).

If this value is left empty then it will fetch all the trades for your account for the given market which may be excessive and can result in your bot hitting or exceeding the rate limit. This configuration helps you to set the starting point from where to fetch trades so that you do not fetch more trades than you need to. 

- **`LAST_TRADE_CURSOR`**: cursor from where to start fetching trades. If left blank then it will fetch from the first trade made on the specified market.

**Note 1**: The `LAST_TRADE_CURSOR` should be specified in the same format as defined by your exchange. On SDEX this can be the paging token, on Kraken this can be your transaction ID, on binance this may be your timestamp etc. You will need to figure out the value to be used. The log files for this trading strategy displays the trades as they happen which includes the trade cursor for each trade entry and can be used to fill in the `LAST_TRADE_CURSOR` value. At each update interval of the bot it logs the currently held value for `LAST_TRADE_CURSOR`, which can also be used to update this configuration value when resuming the bot after it has been paused.

**Note 2**: The first time that the bot fetches trades from the cursor specified in the `LAST_TRADE_CURSOR` at startup, it will update the value held in memory for `LAST_TRADE_CURSOR` but will not use the price from these values retrieved to update the bot's orders because it will use the price set in the `SEED_LAST_TRADE_PRICE` (configured above) for the initial run of the bot. This allows you to set a new price for the bot via the `SEED_LAST_TRADE_PRICE` configuration parameter if you are resuming the bot under new market conditions compared to the last run of your bot. For every subsequent trade it will update the vale of `SEED_LAST_TRADE_PRICE` along with `LAST_TRADE_CURSOR` held in memory. This behavior allows you to leave the `LAST_TRADE_CURSOR` setting as-is if your bot has not seen many trades (i.e. for short runs of the bot). Although, it is highly recommend to always update the value of `LAST_TRADE_CURSOR` in the configuration before starting your bot.

## Comparison to Balanced Strategy

This strategy functions similarly to the [balanced strategy](balanced.md) but gives you the ability to control the order size (amount).

Another benefit of this strategy over the balanced strategy is that you do not need a fixed ratio of balances of your assets to begin trading. The amount and initial price is set in the configuration file directly instead of being computed from the balances of the assets in the account, which makes this strategy more flexible than the balanced strategy.

However, one of the tradeoffs of this additional flexibility is that this strategy can run out of one of the assets. To safeguard from this, you can set up _Price Limits_ and _Minimum Amount Limits_ as described in the configuration sections above.

## Run Kelp

Assuming your botConfig is called `trader.cfg` and your strategy config is called `pendulum.cfg`, you can run `kelp` with the following command:

```
kelp trade --botConf ./path/trader.cfg --strategy pendulum --stratConf ./path/pendulum.cfg
```

# Above and Beyond

You can also play around with the configuration parameters of the [sample configuration file for the pendulum strategy](../../configs/trader/sample_pendulum.cfg), look at some of the other strategies that are available out-of-the-box or dig into the code and _create your own strategy_.
