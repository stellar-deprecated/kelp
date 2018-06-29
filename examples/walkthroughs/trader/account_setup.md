# Account Setup

This guide walks through how you can set up your accounts for making markets using an example token `COUPON` that we want to trade.

## Create Accounts

You will need two accounts:

- Trader Account: this account is where the actual trades originate from, i.e. the "owner" of the trades.
- Source Account: (optional) this account allows you to separate out fee payments and sequence numbers.

Note: Many bots can use the same Trader Account if they use different source accounts.

We are testing on the test network, so we go to the [Test Network on Stellar Laboratory](https://www.stellar.org/laboratory/#account-creator?network=test) to create and fund two accounts. These are the accounts we generated:

| Account        | Secret Key                                               | Public Key                                               |
| -------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| Trader Account | SAOQ6IG2WWDEP47WEJNLIU27OBODMEWFDN6PVUR5KHYDOCVCL34J2CUD | GCB7WIQ3TILJLPOT4E7YMOYF6A5TKYRWK3ZHJ5UR6UKD7D7NJVWNWIQV |
| Source Account | SDDAHRX2JB663N3OLKZIBZPF33ZEKMHARX362S737JEJS2AX3GJZY5LU | GBHXGGUD3LIAWJHFO7737C4TFNDDDLZ74C6VBEPF5H53XNRCVIUWZA5I |

## Create Trustlines

Only the **Trader Account** needs to **trust** the assets that the bot will be trading. We created a trustline from the **Trader Account** (`GCB7WIQ3TILJLPOT4E7YMOYF6A5TKYRWK3ZHJ5UR6UKD7D7NJVWNWIQV`) to the `COUPON` asset issued by `GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI`. You can find this transaction [here](https://horizon-testnet.stellar.org/transactions/288d3ada33fac916b30fadc73d1bf0eacf99d8556a8b4a183dfcc2470e2c05a8).

## Acquire Funds

You will need to acquire the relevant tokens with which to trade. In our example we are trading the **native XLM** against the `COUPON:GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI` token.

We go through the necessary steps required by the issuer to issue us `10,000 COUPON` tokens. In this example, we own the issuer account on the test network so we issued these tokens to our **trader account** which you can find [here](https://horizon-testnet.stellar.org/transactions/b148f207c53049c8a2766f1b6497a847bcea6a9584318f719d561e7168ede74d).

## Config Setup

You will need to set up your `botConfig` file which describes the accounts used by the bot. These are the fields you need to set:

- `TRADING_SECRET_SEED`: Here you can fill in the **secret key** for the `Trading Account`.
- `SOURCE_SECRET_SEED`: If you have specified a `Source Account` you can fill in the **secret key** here.
- `ASSET_CODE_A`: This is the **asset code for the** [**base asset**](https://en.wikipedia.org/wiki/Currency_pair#Base_currency), in our case it's `XLM`.
- `ISSUER_A`: This is the issuer account for the **base asset**. In our case our base asset is the native `XLM` asset so we do not specify this field.
- `ASSET_CODE_B`: This is the **asset code for the** [**quote asset**](https://en.wikipedia.org/wiki/Currency_pair), in our case this is `COUPON`.
- `ISSUER_B`: This is the issuer account for the **quote asset**. In our case this is `GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI`.
- `TICK_INTERVAL_SECONDS`: This describes how often you want the bot to run in seconds. We want the bot to run once every 300 seconds (5 minutes).
- `HORIZON_URL`: This is the **url for your horizon instance**. In our case we want to use the test network so we put in `https://horizon-testnet.stellar.org`.

Note: The bot automatically figures out the `account address` from the `secret key` so you don't need to enter that anywhere.

We've created a sample config file from this setup, which you can [find here](../../configs/trader/sample_botConfig.cfg).

## Next Steps

After completing this guide successfully, you should be in a good position to pick the strategy you want to run. You can find a [walkthrough for a simple strategy here](simple.md)
