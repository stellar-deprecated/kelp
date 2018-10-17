# Account Setup

This guide walks you through step 1 of using Kelp’s trading strategies: setting up your Stellar accounts. In this example walkthrough we use an example token `COUPON` that we eventually want to trade. The strategy walkthrough guides will reference the `COUPON` token. 

## Create Accounts

For the purposes of this walkthrough we will setup two accounts: a **trader account** and a **source account**. Account structure is highly configurable based on your trading needs. Depending on your configuration the source account is optional, read this [document](https://www.stellar.org/developers/guides/channels.html) for more information.

 - Trader Account: the account where the trade originates from (i.e. the "owner" of the trades)
 - Source Account: (optional) this account allows your bot to consume fee payments and sequence numbers from a different account. This is useful if you want multiple bots to use the same trader account.
_Note: Bots can use the same Trader Account if they use different source accounts._

This walkthrough runs on the Stellar Test network. In order to create and fund two accounts we’ll need to go to the [Stellar Laboratory](https://www.stellar.org/laboratory/#account-creator?network=test) where we can create and fund two accounts on the test network using friendbot. Doing so, we generated the following accounts:

| Account        | Secret Key                                               | Public Key                                               |
| -------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| Trader Account | SAOQ6IG2WWDEP47WEJNLIU27OBODMEWFDN6PVUR5KHYDOCVCL34J2CUD | GCB7WIQ3TILJLPOT4E7YMOYF6A5TKYRWK3ZHJ5UR6UKD7D7NJVWNWIQV |
| Source Account | SDDAHRX2JB663N3OLKZIBZPF33ZEKMHARX362S737JEJS2AX3GJZY5LU | GBHXGGUD3LIAWJHFO7737C4TFNDDDLZ74C6VBEPF5H53XNRCVIUWZA5I |

## Create Trustlines

For a Stellar account to trade or hold a particular asset, it must first establish a **trustline** to it. Since only your **Trader Account** will actually trade or hold assets, only it needs trustlines. 

In our case, before we can fund our Trader Account with our example token, `COUPON`, we need to create a trustline to it. The public address of `COUPON`'s issuer is `GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI`.

## Acquire Funds

You will need to acquire the relevant tokens that you want to trade. In our example we are trading the **native XLM** against the `COUPON:GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI` token.

We go through the necessary steps required by the issuer to issue us `10,000 COUPON` tokens. In this example, we own the issuer account on the test network so we issued these tokens to our **trader account**.

## Config Setup

You will need to set up your `trader.cfg` file which describes the accounts used by the bot. You need to set the following fields: 

- `TRADING_SECRET_SEED`: **secret key** for the `Trading Account`.
- `SOURCE_SECRET_SEED`: if you have specified a `Source Account` fill in the **secret key** here.
- `ASSET_CODE_A`: **asset code for the** [**base asset**](https://en.wikipedia.org/wiki/Currency_pair#Base_currency), in our case it's `XLM`.
- `ISSUER_A`: issuer account for the **base asset**. In our case our base asset is the native `XLM` asset so we do not specify this field (this only applies to the native asset, XLM).
- `ASSET_CODE_B`: **asset code for the** [**quote asset**](https://en.wikipedia.org/wiki/Currency_pair), in our case this is `COUPON`.
- `ISSUER_B`: issuer account for the **quote asset**. In our case this is `GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI`.
- `TICK_INTERVAL_SECONDS`: how often you want the bot to run in seconds. In our case the bot runs every 300 seconds (5 minutes).
- `HORIZON_URL`: **url for your horizon instance**. In our case we want to use the test network so we put in `https://horizon-testnet.stellar.org`.

Note: The bot automatically determines the `account address` from the `secret key` so you don't need to enter that anywhere.

We've created a [sample config file](../../configs/trader/sample_trader.cfg) - take a look! 

# Next Steps

After taking the steps above you will be in a good position to pick a strategy and deploy your bot. Try using it to [make a market for a stablecoin](buysell.md) or to [create liquidity for a stellar-based token](balanced.md).
