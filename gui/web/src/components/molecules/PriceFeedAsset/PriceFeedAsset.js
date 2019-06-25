import React, { Component } from 'react';
import PropTypes from 'prop-types';
import PriceFeedTitle from '../PriceFeedTitle/PriceFeedTitle';
import PriceFeedSelector from '../PriceFeedSelector/PriceFeedSelector';
import fetchPrice from '../../../kelp-ops-api/fetchPrice';

const optionsMetadata = {
  type: "dropdown",
  options: {
    "crypto": {
      value: "crypto",
      text: "Crypto from CMC",
      subtype: {
        type: "dropdown",
        options: {
          "https://api.coinmarketcap.com/v1/ticker/stellar/": {
            value: "https://api.coinmarketcap.com/v1/ticker/stellar/",
            text: "Stellar",
            subtype: null,
          },
          "https://api.coinmarketcap.com/v1/ticker/bitcoin/": {
            value: "https://api.coinmarketcap.com/v1/ticker/bitcoin/",
            text: "Bitcoin",
            subtype: null,
          },
          "https://api.coinmarketcap.com/v1/ticker/ethereum/": {
            value: "https://api.coinmarketcap.com/v1/ticker/ethereum/",
            text: "Ethereum",
            subtype: null,
          },
          "https://api.coinmarketcap.com/v1/ticker/litecoin/": {
            value: "https://api.coinmarketcap.com/v1/ticker/litecoin/",
            text: "Litecoin",
            subtype: null,
          },
          "https://api.coinmarketcap.com/v1/ticker/tether/": {
            value: "https://api.coinmarketcap.com/v1/ticker/tether/",
            text: "Tether",
            subtype: null,
          }
        }
      }
    },
    "fiat": {
      value: "fiat",
      text: "Fiat from CurrencyLayer",
      subtype: {
        type: "dropdown",
        options: {
          "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=USD": {
            value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=USD",
            text: "USD",
            subtype: null,
          },
          "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=EUR": {
            value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=EUR",
            text: "EUR",
            subtype: null,
          },
          "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=GBP": {
            value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=GBP",
            text: "GBP",
            subtype: null,
          },
          "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=INR": {
            value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=INR",
            text: "INR",
            subtype: null,
          },
          "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=PHP": {
            value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=PHP",
            text: "PHP",
            subtype: null,
          },
          "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=NGN": {
            value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=NGN",
            text: "NGN",
            subtype: null,
          }
        }
      }
    },
    "fixed": {
      value: "fixed",
      text: "Fixed value",
      subtype: {
        type: "text",
        defaultValue: "1.0",
        subtype: null,
      }
    },
    "exchange": {
      value: "exchange",
      text: "Centralized Exchange",
      subtype: {
        type: "dropdown",
        options: {
          "kraken": {
            value: "kraken",
            text: "Kraken",
            subtype: {
              type: "dropdown",
              options: {
                "XXLM/ZUSD": {
                  value: "XXLM/ZUSD",
                  text: "XLM/USD",
                  subtype: null,
                },
                "XXLM/XXBT": {
                  value: "XXLM/XXBT",
                  text: "XLM/BTC",
                  subtype: null,
                },
                "XXBT/ZUSD": {
                  value: "XXBT/ZUSD",
                  text: "BTC/USD",
                  subtype: null,
                },
                "XETH/ZUSD": {
                  value: "XETH/ZUSD",
                  text: "ETH/USD",
                  subtype: null,
                },
                "XETH/XXBT": {
                  value: "XETH/XXBT",
                  text: "ETH/BTC",
                  subtype: null,
                }
              }
            }
          }
        }
      }
    },
    "sdex": {
      value: "sdex",
      text: "Stellar DEX",
      subtype: {
        type: "text",
        defaultValue: "USD:GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX/XLM:",
        subtype: null,
      }
    }
  }
};

class PriceFeedAsset extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isLoading: true,
      price: null
    };
    this.queryPrice = this.queryPrice.bind(this);

    this._asyncRequests = {};
  }

  static propTypes = {
    baseUrl: PropTypes.string,
    title: PropTypes.string,
    type: PropTypes.string,
    feed_url: PropTypes.string,
    onChange: PropTypes.func,
    onLoadingPrice: PropTypes.func,
    onNewPrice: PropTypes.func,
  };

  componentDidMount() {
    this.queryPrice();
  }

  componentDidUpdate(prevProps) {
    if (
      prevProps.baseUrl !== this.props.baseUrl ||
      prevProps.type !== this.props.type ||
      prevProps.feed_url !== this.props.feed_url
    ) {
      this.queryPrice();
    }
  }

  queryPrice() {
    if (this._asyncRequests["price"]) {
      return
    }
    
    this.setState({
      isLoading: true,
    });
    this.props.onLoadingPrice();

    var _this = this;
    this._asyncRequests["price"] = fetchPrice(this.props.baseUrl, this.props.type, this.props.feed_url).then(resp => {
      _this._asyncRequests["price"] = null;
      let updateStateObj = { isLoading: false };
      if (!resp.error) {
        updateStateObj.price = resp.price
        this.props.onNewPrice(resp.price);
      }

      _this.setState(updateStateObj);
    });
  }

  componentWillUnmount() {
    if (this._asyncRequests["price"]) {
      this._asyncRequests["price"].cancel();
      this._asyncRequests["price"] = null;
    }
  }

  render() {
    let title = (<PriceFeedTitle
      label={this.props.title}
      loading={false}
      price={this.state.price}
      fetchPrice={this.queryPrice}
      />);
    if (this.state.isLoading) {
      title = (<PriceFeedTitle
        label={this.props.title}
        loading={true}
        fetchPrice={this.queryPrice}
        />);
    }

    let values = [this.props.type];
    if (this.props.type === "exchange") {
      let parts = this.props.feed_url.split('/');
      values.push(parts[0]);
      if (parts.length > 1) {
        // then it has to be 3
        values.push(parts[1] + "/" + parts[2]);
      }
    } else {
      values.push(this.props.feed_url);
    }

    return (
      <div>
        {title}
        <PriceFeedSelector
          optionsMetadata={optionsMetadata}
          values={values}
          onChange={this.props.onChange}
          />
      </div>
    );
  }
}

export default PriceFeedAsset;