import React, { Component } from 'react';
import Form from '../../molecules/Form/Form';
import getBotConfig from '../../../kelp-ops-api/getBotConfig';
import getNewBotConfig from '../../../kelp-ops-api/getNewBotConfig';
import upsertBotConfig from '../../../kelp-ops-api/upsertBotConfig';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';

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
          },
          "ccxt-binance": {
            value: "ccxt-binance",
            text: "Binance (via CCXT)",
            subtype: {
              type: "dropdown",
              options: {
                "BTC/USDT": {
                  value: "BTC/USDT",
                  text: "BTC/USDT",
                  subtype: null,
                },
                "ETH/USDT": {
                  value: "ETH/USDT",
                  text: "ETH/USDT",
                  subtype: null,
                },
                "BNB/USDT": {
                  value: "BNB/USDT",
                  text: "BNB/USDT",
                  subtype: null,
                },
                "BNB/BTC": {
                  value: "BNB/USDT",
                  text: "BNB/USDT",
                  subtype: null,
                },
                "XLM/USDT": {
                  value: "XLM/USDT",
                  text: "XLM/USDT",
                  subtype: null,
                },
              }
            }
          }
        }
      }
    // },
    // "sdex": {
    //   value: "sdex",
    //   text: "Stellar DEX",
    //   subtype: {
    //     type: "text",
    //     defaultValue: "USD:GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX/XLM:",
    //     subtype: null,
    //   }
    }
  }
};

class NewBot extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isSaving: false,
      configData: null,
      errorResp: null,
      optionsMetadata: null,
    };

    this.saveNew = this.saveNew.bind(this);
    this.saveEdit = this.saveEdit.bind(this);
    this.loadNewConfigData = this.loadNewConfigData.bind(this);
    this.loadBotConfigData = this.loadBotConfigData.bind(this);
    this.onChangeForm = this.onChangeForm.bind(this);
    this.updateUsingDotNotation = this.updateUsingDotNotation.bind(this);

    this._asyncRequests = {};
  }

  componentWillMount() {
    var _this = this;
    setTimeout(function() {
      _this.setState({
        optionsMetadata: optionsMetadata,
      })
    }, 5000);
  }

  componentWillUnmount() {
    if (this._asyncRequests["botConfig"]) {
      delete this._asyncRequests["botConfig"];
    }
  }

  saveNew() {
    // same behavior for now. this will diverge when we allow for editing the name of a bot
    this.saveEdit();
  }

  saveEdit() {
    this.setState({
      isSaving: true,
    });

    var _this = this;
    this._asyncRequests["botConfig"] = upsertBotConfig(this.props.baseUrl, JSON.stringify(this.state.configData)).then(resp => {
      if (!_this._asyncRequests["botConfig"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["botConfig"];
      _this.setState({
        isSaving: false,
      });

      if (resp["success"]) {
        _this.props.history.goBack();
      } else if (resp["error"]) {
        _this.setState({
          errorResp: resp,
        });
      } else {
        _this.setState({
          errorResp: { error: "Unknown error while attempting to save bot config" },
        });
      }
    });
  }

  loadNewConfigData() {
    var _this = this;
    this._asyncRequests["botConfig"] = getNewBotConfig(this.props.baseUrl).then(resp => {
      if (!_this._asyncRequests["botConfig"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }
      
      delete _this._asyncRequests["botConfig"];
      _this.setState({
        configData: resp,
      });
    });
  }

  loadBotConfigData(botName) {
    var _this = this;
    this._asyncRequests["botConfig"] = getBotConfig(this.props.baseUrl, botName).then(resp => {
      if (!_this._asyncRequests["botConfig"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }
      
      delete _this._asyncRequests["botConfig"];
      _this.setState({
        configData: resp,
      });
    });
  }

  updateUsingDotNotation(obj, path, newValue) {
    // update correct value by converting from dot notation string
    let parts = path.split('.');

    // maintain reference to original object by creating copy
    let current = obj;

    // fetch the object that contains the field we want to update
    for (let i = 0; i < parts.length - 1; i++) {
      current = current[parts[i]];
    }

    // update the field
    current[parts[parts.length-1]] = newValue;
  }

  onChangeForm(statePath, event, mergeUpdateInstructions) {
    // make copy of current state
    let updateJSON = Object.assign({}, this.state);

    if (statePath) {
      this.updateUsingDotNotation(updateJSON.configData, statePath, event.target.value);
    }

    // merge in any additional updates
    if (mergeUpdateInstructions) {
      let keys = Object.keys(mergeUpdateInstructions)
      for (let i = 0; i < keys.length; i++) {
        let dotNotationKey = keys[i];
        let fn = mergeUpdateInstructions[dotNotationKey];
        let newValue = fn(event.target.value);
        if (newValue != null) {
          this.updateUsingDotNotation(updateJSON.configData, dotNotationKey, newValue);
        }
      }
    }

    // set state for the full state object
    this.setState(updateJSON);
  }

  render() {
    if (this.state.isSaving) {
      return (<LoadingAnimation/>);
    }

    if (this.props.location.pathname === "/new") {
      if (!this.state.configData) {
        this.loadNewConfigData();
        return (<div>Fetching sample config file</div>);
      }
      return (<Form
        router={this.props.history}
        isNew={true}
        baseUrl={this.props.baseUrl}
        title="New Bot"
        optionsMetadata={this.state.optionsMetadata}
        onChange={this.onChangeForm}
        configData={this.state.configData}
        saveFn={this.saveNew}
        saveText="Create Bot"
        errorResp={this.state.errorResp}
        />);
    } else if (this.props.location.pathname !== "/edit") {
      console.log("invalid path: " + this.props.location.pathname);
      return "";
    }

    if (this.props.location.search.length === 0) {
      console.log("no search params provided to '/edit' route");
      return "";
    }

    let searchParams = new URLSearchParams(this.props.location.search.substring(1));
    let botNameEncoded = searchParams.get("bot_name");
    if (!botNameEncoded) {
      console.log("no botName param provided to '/edit' route");
      return "";
    }

    let botName = decodeURIComponent(botNameEncoded);
    if (!this.state.configData) {
      this.loadBotConfigData(botName);
      return (<div>Fetching config file for bot: {botName}</div>);
    }
    return (<Form 
      router={this.props.history}
      isNew={false}
      baseUrl={this.props.baseUrl}
      title="Edit Bot"
      optionsMetadata={this.state.optionsMetadata}
      onChange={this.onChangeForm}
      configData={this.state.configData}
      saveFn={this.saveEdit}
      saveText="Save Bot Updates"
      errorResp={this.state.errorResp}
      />);
  }
}

export default NewBot;
