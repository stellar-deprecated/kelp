import React, { Component } from 'react';
import styles from './Form.module.scss';
import grid from '../../_styles/grid.module.scss';
import Badge from '../../atoms/Badge/Badge';
import Input from '../../atoms/Input/Input';
import Label from '../../atoms/Label/Label';
import SectionTitle from '../../atoms/SectionTitle/SectionTitle';
import Switch from '../../atoms/Switch/Switch';
import SegmentedControl from '../../atoms/SegmentedControl/SegmentedControl';
import SectionDescription from '../../atoms/SectionDescription/SectionDescription';
import Button from '../../atoms/Button/Button';
// import Select from '../../atoms/Select/Select';
import FieldItem from '../FieldItem/FieldItem';
import ScreenHeader from '../ScreenHeader/ScreenHeader';
import AdvancedWrapper from '../AdvancedWrapper/AdvancedWrapper';
import FormSection from '../FormSection/FormSection';
import FieldGroup from '../FieldGroup/FieldGroup';
import PriceFeedAsset from '../PriceFeedAsset/PriceFeedAsset';
import FiatFeedAPIKey from '../FiatFeedAPIKey/FiatFeedAPIKey';
import PriceFeedFormula from '../PriceFeedFormula/PriceFeedFormula';
import Levels from '../Levels/Levels';
import ErrorMessage from '../ErrorMessage/ErrorMessage';
import newSecretKey from '../../../kelp-ops-api/newSecretKey';
import fetchPrice from '../../../kelp-ops-api/fetchPrice';
import SecretKey from '../SecretKey/SecretKey';

const fiatURLPrefix = "http://apilayer.net/api/live?access_key=";
const fiatURLCurrencyParam = "&currencies=";
const currencyLayerWebsite = "https://currencylayer.com/";

class Form extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isSaving: false,
      isLoadingFormula: true,
      numerator: null,
      denominator: null,
      numericalErrors: {},
      levelNumericalErrors: {},
      attemptedFirstSave: false,
      fiatAPIKey: this._extractFiatAPIKey(props),
    };
    this.setLoadingFormula = this.setLoadingFormula.bind(this);
    this.updateFormulaPrice = this.updateFormulaPrice.bind(this);
    this.fetchErrorMessage = this.fetchErrorMessage.bind(this);
    this.hasPriceFeedError = this.hasPriceFeedError.bind(this);
    this.fiatAPIKeyError = this.fiatAPIKeyError.bind(this);
    this.save = this.save.bind(this);
    this.priceFeedAssetChangeHandler = this.priceFeedAssetChangeHandler.bind(this);
    this.updateLevel = this.updateLevel.bind(this);
    this.newLevel = this.newLevel.bind(this);
    this.hasNewLevel = this.hasNewLevel.bind(this);
    this.removeLevel = this.removeLevel.bind(this);
    this.newSecret = this.newSecret.bind(this);
    this.getError = this.getError.bind(this);
    this.addNumericalError = this.addNumericalError.bind(this);
    this.clearNumericalError = this.clearNumericalError.bind(this);
    this.getNumNumericalErrors = this.getNumNumericalErrors.bind(this);
    this.addLevelError = this.addLevelError.bind(this);
    this.clearLevelError = this.clearLevelError.bind(this);
    this.makeNewFiatDataFeedURL = this.makeNewFiatDataFeedURL.bind(this);
    this.extractCurrencyCodeFromFiatURL = this.extractCurrencyCodeFromFiatURL.bind(this);
    this.updateFiatAPIKey = this.updateFiatAPIKey.bind(this);
    this.getConfigFeedURLTransformIfFiat = this.getConfigFeedURLTransformIfFiat.bind(this);
    this._emptyLevel = this._emptyLevel.bind(this);
    this._triggerUpdateLevels = this._triggerUpdateLevels.bind(this);
    this._fetchDotNotation = this._fetchDotNotation.bind(this);
    this._last_fill_tracker_sleep_millis = 1000;

    this._asyncRequests = {};
  }

  // _extractFiatAPIKey gets called when we load the config and we want to populate the fiat APIKey in the GUI
  _extractFiatAPIKey(props) {
    let url = null;
    if (props.configData.strategy_config.data_type_a === "fiat") {
      const urlA = props.configData.strategy_config.data_feed_a_url;
      if (urlA.startsWith(fiatURLPrefix) && urlA.indexOf(fiatURLCurrencyParam) > 0) {
        url = urlA;
      }
    } else if (props.configData.strategy_config.data_type_b === "fiat") {
      const urlB = props.configData.strategy_config.data_feed_b_url;
      if (urlB.startsWith(fiatURLPrefix) && urlB.indexOf(fiatURLCurrencyParam) > 0) {
        url = urlB;
      }
    }

    if (!url) {
      return "";
    }
    return url.substring(fiatURLPrefix.length, url.indexOf(fiatURLCurrencyParam));
  }

  // getConfigFeedURLTransformIfFiat is called when loading the config value for feed URLs
  // we want to load only the currencyCode
  getConfigFeedURLTransformIfFiat(ab) {
    let dataType = this.props.configData.strategy_config["data_type_" + ab]
    let feedUrl = this.props.configData.strategy_config["data_feed_" + ab + "_url"];

    if (dataType === "fiat") {
      const currencyCode = this.extractCurrencyCodeFromFiatURL(feedUrl);
      return currencyCode;
    }
    return feedUrl;
  }

  componentWillUnmount() {
    if (this._asyncRequests["secretKey"]) {
      delete this._asyncRequests["secretKey"];
    }
  }

  setLoadingFormula() {
    this.setState({
      isLoadingFormula: true
    })
  }

  updateFormulaPrice(ator, price) {
    const isPriceMissing = price === null || price === undefined;
    if (ator === "numerator") {
      this.setState({
        isLoadingFormula: isPriceMissing || this.state.denominator === null || this.state.denominator === undefined,
        numerator: price
      })
    } else {
      this.setState({
        isLoadingFormula: isPriceMissing || this.state.numerator === null || this.state.numerator === undefined,
        denominator: price
      })
    }
  }

  _fetchDotNotation(obj, path) {
    if (obj === undefined || obj === null || obj === "" || obj === 0) {
      return null;
    }
    
    let parts = path.split('.');
    for (let i = 0; i < parts.length; i++) {
      obj = obj[parts[i]];

      if (obj === undefined || obj === null || obj === "" || obj === 0) {
        return null;
      }
    }

    return obj
  }

  getError(fieldKey) {
    let numericalError = this.state.numericalErrors[fieldKey];
    if (numericalError) {
      return numericalError;
    }

    if (!this.props.errorResp) {
      return null;
    }

    return this._fetchDotNotation(this.props.errorResp.fields, fieldKey);
  }

  getNumNumericalErrors() {
    return Object.keys(this.state.numericalErrors).length;
  }

  hasPriceFeedError() {
    return isNaN(this.state.numerator) || isNaN(this.state.denominator) || this.state.numerator < 0 || this.state.denominator < 0;
  }

  fiatAPIKeyError() {
    const hasNumeratorFiatFeedWithError = this.props.configData.strategy_config.data_type_a === "fiat" && this.state.numerator && this.state.numerator < 0;
    const hasDenominatorFiatFeedWithError = this.props.configData.strategy_config.data_type_b === "fiat" && this.state.denominator && this.state.denominator < 0;
    if (hasNumeratorFiatFeedWithError || hasDenominatorFiatFeedWithError) {
      return "invalid API key, exhausted API limit, or API account inactive. Go to " + currencyLayerWebsite + " to sign up for an API key."
    }
    return null;
  }

  save() {
    if (this.getNumNumericalErrors() > 0) {
      // set state so we refresh and show the error message
      this.setState({
        isSaving: false,
        attemptedFirstSave: true,
      });
      return;
    }

    if (this.hasPriceFeedError()) {
      // set state so we refresh and show the error message
      this.setState({
        isSaving: false,
        attemptedFirstSave: true,
      });
      return;
    }

    this.setState({
      isSaving: true,
      attemptedFirstSave: true,
    })
    this.props.saveFn();
    // save fn will call router.goBack();
  }

  priceFeedAssetChangeHandler(ab, newValues) {
    let dataTypeFieldName = "strategy_config.data_type_" + ab;
    let dataTypeValue = newValues[0];
    let feedUrlFieldName = "strategy_config.data_feed_" + ab + "_url";
    // special handling for feedUrlValue
    let feedUrlValue = newValues[1];
    if (newValues.length > 2) {
      feedUrlValue = feedUrlValue + "/" + newValues[2];
    }

    // when the users selects a new fiat feed currency, wrap the currencyCode (feedUrlValue) with the fiat URL format
    // so it can be saved in the config file which requires a URL.
    // This leaves the UI text in the dropdown unchanged
    if (dataTypeValue === "fiat") {
      feedUrlValue = this.makeNewFiatDataFeedURL(this.state.fiatAPIKey, feedUrlValue);
    }

    let mergeUpdateInstructions = {};
    mergeUpdateInstructions[feedUrlFieldName] = () => { return feedUrlValue };
    this.props.onChange(dataTypeFieldName, {target: {value: dataTypeValue}}, mergeUpdateInstructions);
  }

  updateLevel(levelIdx, fieldAmtSpread, value) {
    let newLevels = this.props.configData.strategy_config.levels;
    newLevels[levelIdx][fieldAmtSpread] = value;
    this._triggerUpdateLevels(newLevels);
  }

  addLevelError(levelIdx, subfield, message) {
    levelIdx = "" + levelIdx;
    
    let newLevelNumericalErrors = this.state.levelNumericalErrors;
    let newValue = newLevelNumericalErrors[levelIdx];
    if (!newValue) {
      newValue = {
        spread: null,
        amount: null,
      };
    }
    newValue[subfield] = message;
    newLevelNumericalErrors[levelIdx] = newValue;
    this.setState({
      levelNumericalErrors: newLevelNumericalErrors
    });

    this.addNumericalError("strategy_config.levels", "there is an error in one of the levels");
  }
  
  clearLevelError(levelIdx, subfield) {
    levelIdx = "" + levelIdx;

    let newLevelNumericalErrors = this.state.levelNumericalErrors;
    let newValue = newLevelNumericalErrors[levelIdx];
    if (newValue) {
      newValue[subfield] = null;
      if (!newValue.spread && !newValue.amount) {
        delete newLevelNumericalErrors[levelIdx];
      } else {
        newLevelNumericalErrors[levelIdx] = newValue;
        this.setState({
          levelNumericalErrors: newLevelNumericalErrors
        });
      }
    }

    if (Object.keys(newLevelNumericalErrors).length === 0) {
      this.clearNumericalError("strategy_config.levels")
    }
  }

  hasNewLevel() {
    let levels = this.props.configData.strategy_config.levels;
    if (levels.length === 0) {
      return false;
    }

    let lastLevel = levels[levels.length - 1];
    if (+lastLevel["spread"] === 0 || +lastLevel["amount"] === 0) {
      return true;
    }

    return false;
  }

  newLevel() {
    if (this.hasNewLevel()) {
      return
    }

    let newLevels = this.props.configData.strategy_config.levels;

    // push default values for new level
    newLevels.push(this._emptyLevel());

    this._triggerUpdateLevels(newLevels);
  }

  removeLevel(levelIdx) {
    if (levelIdx >= this.props.configData.strategy_config.levels.length) {
      return;
    }

    let newLevels = this.props.configData.strategy_config.levels.filter((_, idx) => idx !== levelIdx);
    if (newLevels.length === 0) {
      newLevels.push(this._emptyLevel());
    }

    this.clearLevelError(levelIdx, "spread");
    this.clearLevelError(levelIdx, "amount");

    this._triggerUpdateLevels(newLevels);
  }

  _emptyLevel() {
    return {
      amount: 0.00,
      spread: 0.00,
    };
  }

  _triggerUpdateLevels(newLevels) {
    // update levels and always set amount_of_a_base to 1.0
    this.props.onChange(
      "strategy_config.levels", {target: {value: newLevels}},
      { "strategy_config.amount_of_a_base": (value) => { return 1.0; } }
    )
  }

  newSecret(field) {
    var _this = this;
    this._asyncRequests["secretKey"] = newSecretKey(this.props.baseUrl).then(resp => {
      if (!_this._asyncRequests["secretKey"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["secretKey"];
      this.props.onChange(field, {target: {value: resp}});
    });
  }

  makeNewFiatDataFeedURL(apiKey, currencyCode) {
    return fiatURLPrefix + apiKey + fiatURLCurrencyParam + currencyCode;
  }

  extractCurrencyCodeFromFiatURL(url) {
    return url.substring(url.indexOf(fiatURLCurrencyParam) + fiatURLCurrencyParam.length);
  }

  // updateFiatAPIKey is called when the user upates the fiat API key in the GUI
  updateFiatAPIKey(apiKey) {
    if (this.props.configData.strategy_config.data_type_a === "fiat") {
      const newValue = this.makeNewFiatDataFeedURL(apiKey, this.extractCurrencyCodeFromFiatURL(this.props.configData.strategy_config.data_feed_a_url));
      this.props.onChange("strategy_config.data_feed_a_url", { target: { value: newValue } });
    }

    if (this.props.configData.strategy_config.data_type_b === "fiat") {
      const newValue = this.makeNewFiatDataFeedURL(apiKey, this.extractCurrencyCodeFromFiatURL(this.props.configData.strategy_config.data_feed_b_url));
      this.props.onChange("strategy_config.data_feed_b_url", { target: { value: newValue } });
    }

    this.setState({
      fiatAPIKey: apiKey,
    })
  }

  clearNumericalError(field) {
    let newNumericalErrors = this.state.numericalErrors;
    delete newNumericalErrors[field];
    this.setState({
      numericalErrors: newNumericalErrors
    });
  }

  addNumericalError(field, message) {
    let newNumericalErrors = this.state.numericalErrors;
    newNumericalErrors[field] = message;
    this.setState({
      numericalErrors: newNumericalErrors
    });
  }

  fetchErrorMessage() {
    let errorList = [];
    if (this.props.errorResp) {
      errorList.push(this.props.errorResp.error);
    }
    if (this.getNumNumericalErrors() > 0) {
      errorList.push("there are some invalid numerical values inline");
    }
    if (this.state.attemptedFirstSave && this.hasPriceFeedError()) {
      errorList.push("the computed price feed is invalid");
    }

    let error = "";
    if (errorList.length > 0) {
      error = (<ErrorMessage errorList={errorList}/>);
    }
    return error;
  }

  render() {
    // let tradingPlatform = "sdex";
    // if (this.props.configData.trader_config.trading_exchange && this.props.configData.trader_config.trading_exchange !== "") {
    //   tradingPlatform = this.props.configData.trader_config.trading_exchange;
    // }

    let isTestNet = this.props.configData.trader_config.horizon_url.includes("test");
    let network = "PubNet";
    if (isTestNet) {
      network = "TestNet";
    }

    const error = this.fetchErrorMessage();

    let errorSubmitContainer = (
      <div className={grid.container}>
        {error}
        <div className={styles.formFooter}>
          <Button
            eventName={this.props.eventPrefix + "-save"}
            icon="add" 
            size="large" 
            loading={this.state.isSaving}
            disabled={this.props.optionsMetadata == null}
            onClick={this.save}
            >
            {this.props.saveText}
          </Button>
        </div>
      </div>
    );
    if (this.props.readOnly) {
      errorSubmitContainer = "";
    }

    let readOnlyMessage = "";
    if (this.props.readOnly) {
      readOnlyMessage = (<Badge type="message" value="read only"/>);
    }

    return (
      <div>
        <div className={grid.container}>
            <ScreenHeader
              title={this.props.title}
              backButtonFn={this.props.router.goBack}
              eventPrefix={this.props.eventPrefix}
              >
                {/* <Switch/>
                <Label>Helper Fields</Label> */}
                {readOnlyMessage}
            </ScreenHeader>

            {error}

            <FormSection>
              <Input
                size="large"
                value={this.props.configData.name}
                type="string"
                onChange={(event) => { this.props.onChange("name", event) }}
                disabled={!this.props.isNew}
                error={this.getError("name")}
                readOnly={this.props.readOnly}
                />

              {/* Trader Settings */}
              <SectionTitle>
                Trader Settings
              </SectionTitle>
            </FormSection>
            
            <FormSection>
              <SectionDescription>
                These settings refer to the operations of the bot, trading platform, and runtime parameters, but not the chosen trading strategy.
                <br/>
                <br/>
                Scroll below to see the strategy settings.
              </SectionDescription>
            </FormSection>

            {/* <FormSection tip="Where do you want to trade: Stellar Decentralized Exchange (SDEX) or Kraken?">
              <FieldItem>
                <Label>Trading Platform</Label>
                <Select
                  options={[
                      {value: "sdex", text: "SDEX"},
                      {value: "kraken", text: "Kraken"},
                    ]}
                  selected={tradingPlatform}
                  />
              </FieldItem>
            </FormSection> */}
              
            <FormSection>
            <FieldItem>
              <Label padding>Network</Label>
              <SegmentedControl
                segments={this.props.segmentNetworkOptions}
                selected={network}
                onSelect={(selected) => {
                  // TODO use URI passed in from command line, or indicate to backend it's test/public
                  let newValue = "https://horizon-testnet.stellar.org";
                  if (selected === "PubNet") {
                    newValue = "https://horizon.stellar.org";
                  }
                  this.props.onChange("trader_config.horizon_url", { target: { value: newValue } });
                }}
                error={this.getError("trader_config.horizon_url")}
              />
            </FieldItem>
          </FormSection>
            
            <FormSection>
              <FieldItem>
                <SecretKey
                  label="Trader account secret key"
                  isTestNet={isTestNet}
                  secret={this.props.configData.trader_config.trading_secret_seed}
                  onSecretChange={(event) => { this.props.onChange("trader_config.trading_secret_seed", event) }}
                  onError={() => this.getError("trader_config.trading_secret_seed")}
                  onNewKeyClick={() => this.newSecret("trader_config.trading_secret_seed")}
                  readOnly={this.props.readOnly}
                  eventPrefix={this.props.eventPrefix + "-secretkey-trader"}
                />
              </FieldItem>
            </FormSection>

            {/* <FormSection tip="Lorem ipsum dolor sit amet, consectetur adipiscing elit.  Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl."> */}
            <FormSection wideCol={80}>
              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldItem>
                    <Label>Base asset code</Label>
                    <Input
                      value={this.props.configData.trader_config.asset_code_a}
                      type="string"
                      placeholder="XLM"
                      onChange={(event) => {
                        this.props.onChange("trader_config.asset_code_a", event, {
                          "trader_config.issuer_a": (value) => {
                            if (value === "XLM") {
                              return "";
                            }
                            return null;
                          }
                        })
                      }}
                      error={this.getError("trader_config.asset_code_a")}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>

                <div className={grid.col8}>
                  <FieldItem>
                    <Label>Base asset issuer</Label>
                    <Input
                      value={this.props.configData.trader_config.issuer_a}
                      type="string"
                      onChange={(event) => { this.props.onChange("trader_config.issuer_a", event) }}
                      disabled={this.props.configData.trader_config.asset_code_a === "XLM"}
                      error={this.getError("trader_config.issuer_a")}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
              </div>

              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldItem>
                    <Label>Quote asset code</Label>
                    <Input
                      value={this.props.configData.trader_config.asset_code_b}
                      type="string"
                      placeholder="COUPON"
                      onChange={(event) => {
                        this.props.onChange("trader_config.asset_code_b", event, {
                          "trader_config.issuer_b": (value) => {
                            if (value === "XLM") {
                              return "";
                            }
                            return null;
                          }
                        })
                      }}
                      error={this.getError("trader_config.asset_code_b")}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>

                <div className={grid.col8}>
                  <FieldItem>
                    <Label>Quote asset issuer</Label>
                    <Input
                      value={this.props.configData.trader_config.issuer_b}
                      type="string"
                      placeholder="GBMMZMK2DC4FFP4CAI6KCVNCQ7WLO5A7DQU7EC7WGHRDQBZB763X4OQI"
                      onChange={(event) => { this.props.onChange("trader_config.issuer_b", event) }}
                      disabled={this.props.configData.trader_config.asset_code_b === "XLM"}
                      error={this.getError("trader_config.issuer_b")}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
              </div>
            </FormSection>
        </div>

        <AdvancedWrapper headerClass={grid.container} isOpened={this.props.readOnly}>
          <div className={grid.container}>
            <FormSection>
              <FieldItem>
                <SecretKey
                  label="Source account secret key"
                  isTestNet={isTestNet}
                  secret={this.props.configData.trader_config.source_secret_seed}
                  onSecretChange={(event) => { this.props.onChange("trader_config.source_secret_seed", event) }}
                  onError={() => this.getError("trader_config.source_secret_seed")}
                  onNewKeyClick={() => this.newSecret("trader_config.source_secret_seed")}
                  optional={true}
                  readOnly={this.props.readOnly}
                  eventPrefix={this.props.eventPrefix + "-secretkey-source"}
                />
              </FieldItem>

              <FieldItem>
                <Label>Tick interval</Label>
                <Input
                  suffix="seconds"
                  value={this.props.configData.trader_config.tick_interval_seconds}
                  type="int_positive"
                  onChange={(event) => { this.props.onChange("trader_config.tick_interval_seconds", event) }}
                  error={this.getError("trader_config.tick_interval_seconds")}
                  triggerError={(message) => { this.addNumericalError("trader_config.tick_interval_seconds", message) }}
                  clearError={() => { this.clearNumericalError("trader_config.tick_interval_seconds") }}
                  readOnly={this.props.readOnly}
                  />
              </FieldItem>

              <FieldItem>
                <Label>Maximum Randomized Tick Interval Delay</Label>
                <Input
                  suffix="miliseconds"
                  value={this.props.configData.trader_config.max_tick_delay_millis}
                  type="int_nonnegative"
                  onChange={(event) => { this.props.onChange("trader_config.max_tick_delay_millis", event) }}
                  error={this.getError("trader_config.max_tick_delay_millis")}
                  triggerError={(message) => { this.addNumericalError("trader_config.max_tick_delay_millis", message) }}
                  clearError={() => { this.clearNumericalError("trader_config.max_tick_delay_millis") }}
                  readOnly={this.props.readOnly}
                  />
              </FieldItem>

              <FieldItem inline>
                <Switch
                  value={this.props.configData.trader_config.submit_mode === "maker_only"}
                  onChange={(event) => {
                      let newValue = "maker_only"
                      if (this.props.configData.trader_config.submit_mode === "maker_only") {
                        newValue = "both"
                      }
                      this.props.onChange("trader_config.submit_mode", {target: {value: newValue}});
                    }
                  }
                  readOnly={this.props.readOnly}
                  />
                <Label>Maker only mode</Label>  
              </FieldItem>

              <FieldItem>
                <Label>Delete cycles threshold</Label>
                <Input
                  value={this.props.configData.trader_config.delete_cycles_threshold}
                  type="int_nonnegative"
                  onChange={(event) => { this.props.onChange("trader_config.delete_cycles_threshold", event) }}
                  error={this.getError("trader_config.delete_cycles_threshold")}
                  triggerError={(message) => { this.addNumericalError("trader_config.delete_cycles_threshold", message) }}
                  clearError={() => { this.clearNumericalError("trader_config.delete_cycles_threshold") }}
                  readOnly={this.props.readOnly}
                  />
              </FieldItem>

              <FieldItem inline>
                <Switch
                  value={this.props.configData.trader_config.fill_tracker_sleep_millis !== 0}
                  onChange={(event) => {
                      let newValue = 0;
                      if (this.props.configData.trader_config.fill_tracker_sleep_millis === 0) {
                        newValue = this._last_fill_tracker_sleep_millis;
                      }
                      this.props.onChange("trader_config.fill_tracker_sleep_millis", {target: {value: newValue}});
                    }
                  }
                  readOnly={this.props.readOnly}
                  />
                <Label>Fill tracker</Label>
              </FieldItem>

              <FieldItem>
                <Label disabled={this.props.configData.trader_config.fill_tracker_sleep_millis === 0}>Fill tracker duration</Label>
                <Input
                  suffix="miliseconds"
                  value={this.props.configData.trader_config.fill_tracker_sleep_millis === 0 ? this._last_fill_tracker_sleep_millis : this.props.configData.trader_config.fill_tracker_sleep_millis}
                  type="int_positive"
                  disabled={this.props.configData.trader_config.fill_tracker_sleep_millis === 0}
                  error={this.getError("trader_config.fill_tracker_sleep_millis")}
                  onChange={(event) => {
                      this.props.onChange("trader_config.fill_tracker_sleep_millis", event, {
                        "trader_config.fill_tracker_sleep_millis": (value) => {
                          // cannot set value to 0 or empty
                          if (value === 0) {
                            return this._last_fill_tracker_sleep_millis;
                          }

                          // create a side-effect: just save the last value here and don't update the state in this function
                          this._last_fill_tracker_sleep_millis = value;
                          return null;
                        }
                      })
                    }
                  }
                  readOnly={this.props.readOnly}
                  triggerError={(message) => { this.addNumericalError("trader_config.fill_tracker_sleep_millis", message) }}
                  clearError={() => { this.clearNumericalError("trader_config.fill_tracker_sleep_millis") }}
                  />
              </FieldItem>

              <FieldItem>
                <Label disabled={this.props.configData.trader_config.fill_tracker_sleep_millis === 0}>Fill tracker delete cycles threshold</Label>
                <Input
                  value={this.props.configData.trader_config.fill_tracker_delete_cycles_threshold}
                  type="int_nonnegative"
                  disabled={this.props.configData.trader_config.fill_tracker_sleep_millis === 0}
                  readOnly={this.props.readOnly}
                  onChange={(event) => { this.props.onChange("trader_config.fill_tracker_delete_cycles_threshold", event) }}
                  error={this.getError("trader_config.fill_tracker_delete_cycles_threshold")}
                  triggerError={(message) => { this.addNumericalError("trader_config.fill_tracker_delete_cycles_threshold", message) }}
                  clearError={() => { this.clearNumericalError("trader_config.fill_tracker_delete_cycles_threshold") }}
                  />
              </FieldItem>

              <FieldGroup groupTitle="Fee">
                <div className={grid.row}>
                  <div className={grid.col5}>
                    <FieldItem>
                      <Label>Network capacity trigger</Label>
                      <Input
                        value={this.props.configData.trader_config.fee.capacity_trigger}
                        type="float_positive"
                        onChange={(event) => { this.props.onChange("trader_config.fee.capacity_trigger", event) }}
                        error={this.getError("trader_config.fee.capacity_trigger")}
                        triggerError={(message) => { this.addNumericalError("trader_config.fee.capacity_trigger", message) }}
                        clearError={() => { this.clearNumericalError("trader_config.fee.capacity_trigger") }}
                        readOnly={this.props.readOnly}
                        />
                    </FieldItem>
                  </div>
                </div>
                <div className={grid.row}>
                  <div className={grid.col5}>
                    <FieldItem>
                      <Label>Use fee percentile value</Label>
                      <Input
                        suffix="%"
                        value={this.props.configData.trader_config.fee.percentile}
                        type="int_positive"  // this is a selection from the fee stats endpoint
                        onChange={(event) => { this.props.onChange("trader_config.fee.percentile", event) }}
                        error={this.getError("trader_config.fee.percentile")}
                        triggerError={(message) => { this.addNumericalError("trader_config.fee.percentile", message) }}
                        clearError={() => { this.clearNumericalError("trader_config.fee.percentile") }}
                        readOnly={this.props.readOnly}
                        />
                    </FieldItem>
                    </div>
                </div>
                <div className={grid.row}>
                  <div className={grid.col5}>
                    <FieldItem>
                      <Label>Maximum fee per operation</Label>
                      <Input
                        suffix="stroops"
                        value={this.props.configData.trader_config.fee.max_op_fee_stroops}
                        type="int_positive"
                        onChange={(event) => { this.props.onChange("trader_config.fee.max_op_fee_stroops", event) }}
                        error={this.getError("trader_config.fee.max_op_fee_stroops")}
                        triggerError={(message) => { this.addNumericalError("trader_config.fee.max_op_fee_stroops", message) }}
                        clearError={() => { this.clearNumericalError("trader_config.fee.max_op_fee_stroops") }}
                        readOnly={this.props.readOnly}
                        />
                    </FieldItem>
                    </div>
                </div>
              </FieldGroup>

              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Precision units for price</Label>
                    <Input
                      suffix="decimals"
                      value={this.props.configData.trader_config.centralized_price_precision_override}
                      type="int_positive"
                      onChange={(event) => { this.props.onChange("trader_config.centralized_price_precision_override", event) }}
                      error={this.getError("trader_config.centralized_price_precision_override")}
                      triggerError={(message) => { this.addNumericalError("trader_config.centralized_price_precision_override", message) }}
                      clearError={() => { this.clearNumericalError("trader_config.centralized_price_precision_override") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Precision units for volume</Label>
                    <Input
                      suffix="decimals"
                      value={this.props.configData.trader_config.centralized_volume_precision_override}
                      type="int_positive"
                      onChange={(event) => { this.props.onChange("trader_config.centralized_volume_precision_override", event) }}
                      error={this.getError("trader_config.centralized_volume_precision_override")}
                      triggerError={(message) => { this.addNumericalError("trader_config.centralized_volume_precision_override", message) }}
                      clearError={() => { this.clearNumericalError("trader_config.centralized_volume_precision_override") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
              </div>

              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Min. order size (volume) of base units</Label>
                    <Input
                      suffix="units"
                      value={this.props.configData.trader_config.centralized_min_base_volume_override}
                      type="float_positive"
                      onChange={(event) => { this.props.onChange("trader_config.centralized_min_base_volume_override", event) }}
                      error={this.getError("trader_config.centralized_min_base_volume_override")}
                      triggerError={(message) => { this.addNumericalError("trader_config.centralized_min_base_volume_override", message) }}
                      clearError={() => { this.clearNumericalError("trader_config.centralized_min_base_volume_override") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Min. order size (volume) of quote units</Label>
                    <Input
                      suffix="units"
                      value={this.props.configData.trader_config.centralized_min_quote_volume_override}
                      type="float_positive"
                      onChange={(event) => { this.props.onChange("trader_config.centralized_min_quote_volume_override", event) }}
                      error={this.getError("trader_config.centralized_min_quote_volume_override")}
                      triggerError={(message) => { this.addNumericalError("trader_config.centralized_min_quote_volume_override", message) }}
                      clearError={() => { this.clearNumericalError("trader_config.centralized_min_quote_volume_override") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
              </div>  

            </FormSection>
          </div>
        </AdvancedWrapper>


        {/* Stratefy Settings */}
        <div className={grid.container}>
          <FormSection wideCol={70}>
            <SectionTitle>
              Strategy Settings <i>(buysell)</i>
            </SectionTitle>
            
            <SectionDescription>
              These settings refer to the strategy settings for the <i>buysell</i> strategy.
            </SectionDescription>
            

            <FieldGroup groupTitle="Price Feed">
              <FieldItem>
                <PriceFeedAsset
                  onChange={(newValues) => this.priceFeedAssetChangeHandler("a", newValues)}
                  title={"Numerator: current price of base asset (" + this.props.configData.trader_config.asset_code_a + ")"}
                  optionsMetadata={this.props.optionsMetadata}
                  type={this.props.configData.strategy_config.data_type_a}
                  feed_url={this.getConfigFeedURLTransformIfFiat("a")}
                  fetchPrice={fetchPrice.bind(
                    this,
                    this.props.baseUrl,
                    this.props.configData.strategy_config.data_type_a,
                    this.props.configData.strategy_config.data_feed_a_url,
                  )}
                  onLoadingPrice={() => this.setLoadingFormula()}
                  onNewPrice={(newPrice) => this.updateFormulaPrice("numerator", newPrice)}
                  readOnly={this.props.readOnly}
                  eventPrefix={this.props.eventPrefix + "-pricefeed-numerator"}
                  />
              </FieldItem>
              <FieldItem>
                <PriceFeedAsset
                  onChange={(newValues) => this.priceFeedAssetChangeHandler("b", newValues)}
                  title={"Denominator: current price of quote asset (" + this.props.configData.trader_config.asset_code_b + ")"}
                  optionsMetadata={this.props.optionsMetadata}
                  type={this.props.configData.strategy_config.data_type_b}
                  feed_url={this.getConfigFeedURLTransformIfFiat("b")}
                  fetchPrice={fetchPrice.bind(
                    this,
                    this.props.baseUrl,
                    this.props.configData.strategy_config.data_type_b,
                    this.props.configData.strategy_config.data_feed_b_url,
                  )}
                  onLoadingPrice={() => this.setLoadingFormula()}
                  onNewPrice={(newPrice) => this.updateFormulaPrice("denominator", newPrice)}
                  readOnly={this.props.readOnly}
                  eventPrefix={this.props.eventPrefix + "-pricefeed-denominator"}
                  />
              </FieldItem>
              <FieldItem>
                <FiatFeedAPIKey
                  enabled={this.props.configData.strategy_config.data_type_a === "fiat" || this.props.configData.strategy_config.data_type_b === "fiat"}
                  value={this.state.fiatAPIKey}
                  error={this.fiatAPIKeyError()}
                  onChange={(event) => { this.updateFiatAPIKey(event.target.value) }}
                  readOnly={this.props.readOnly}
                  />
              </FieldItem>
              <PriceFeedFormula
                isLoading={this.state.isLoadingFormula || this.props.optionsMetadata == null}
                baseCode={this.props.configData.trader_config.asset_code_a}
                baseIssuer={this.props.configData.trader_config.issuer_a}
                quoteCode={this.props.configData.trader_config.asset_code_b}
                quoteIssuer={this.props.configData.trader_config.issuer_b}
                numerator={this.state.numerator}
                denominator={this.state.denominator}
                />
            </FieldGroup>
            
            <div className={grid.row}>
              <div className={grid.col8}>
                <FieldGroup groupTitle="Levels">
                  <Levels
                    levels={this.props.configData.strategy_config.levels}
                    updateLevel={(levelIdx, subfield, value) => { this.updateLevel(levelIdx, subfield, value) }}
                    newLevel={this.newLevel}
                    hasNewLevel={this.hasNewLevel}
                    onRemove={(levelIdx) => { this.removeLevel(levelIdx) }}
                    error={this.getError("strategy_config.levels")}
                    levelErrors={this.state.levelNumericalErrors}
                    addLevelError={(levelIdx, subfield, message) => { this.addLevelError(levelIdx, subfield, message) }}
                    clearLevelError={(levelIdx, subfield) => { this.clearLevelError(levelIdx, subfield) }}
                    readOnly={this.props.readOnly}
                    eventPrefix={this.props.eventPrefix + "-levels"}
                    />
                </FieldGroup>
              </div>
            </div>
          </FormSection>
        </div>

        <AdvancedWrapper headerClass={grid.container} isOpened={this.props.readOnly}>
          <div className={grid.container}>
            <FormSection>
              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Price tolerance</Label>
                    <Input
                      suffix="%"
                      value={this.props.configData.strategy_config.price_tolerance}
                      type="percent_positive"
                      onChange={(event) => { this.props.onChange("strategy_config.price_tolerance", event) }}
                      error={this.getError("strategy_config.price_tolerance")}
                      triggerError={(message) => { this.addNumericalError("strategy_config.price_tolerance", message) }}
                      clearError={() => { this.clearNumericalError("strategy_config.price_tolerance") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Amount tolerance</Label>
                    <Input
                      suffix="%"
                      value={this.props.configData.strategy_config.amount_tolerance}
                      type="percent_positive"
                      onChange={(event) => { this.props.onChange("strategy_config.amount_tolerance", event) }}
                      error={this.getError("strategy_config.amount_tolerance")}
                      triggerError={(message) => { this.addNumericalError("strategy_config.amount_tolerance", message) }}
                      clearError={() => { this.clearNumericalError("strategy_config.amount_tolerance") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
              </div>

              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Rate offset percentage</Label>
                    <Input
                      suffix="%"
                      value={this.props.configData.strategy_config.rate_offset_percent}
                      type="percent"
                      onChange={(event) => { this.props.onChange("strategy_config.rate_offset_percent", event) }}
                      error={this.getError("strategy_config.rate_offset_percent")}
                      triggerError={(message) => { this.addNumericalError("strategy_config.rate_offset_percent", message) }}
                      clearError={() => { this.clearNumericalError("strategy_config.rate_offset_percent") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Rate offset</Label>
                    <Input
                      value={this.props.configData.strategy_config.rate_offset}
                      type="float"
                      onChange={(event) => { this.props.onChange("strategy_config.rate_offset", event) }}
                      error={this.getError("strategy_config.rate_offset")}
                      triggerError={(message) => { this.addNumericalError("strategy_config.rate_offset", message) }}
                      clearError={() => { this.clearNumericalError("strategy_config.rate_offset") }}
                      readOnly={this.props.readOnly}
                      />
                  </FieldItem>
                </div>
              </div>
              <FieldItem inline>
                <Switch
                  value={this.props.configData.strategy_config.rate_offset_percent_first}
                  onChange={(event) => {
                      let newValue = true;
                      if (this.props.configData.strategy_config.rate_offset_percent_first) {
                        newValue = false;
                      }
                      this.props.onChange("strategy_config.rate_offset_percent_first", {target: {value: newValue}});
                    }
                  }
                  readOnly={this.props.readOnly}
                  />
                <Label>Rate Offset Percent first</Label>
              </FieldItem>
            </FormSection>
          </div>
        </AdvancedWrapper>

        {errorSubmitContainer}      
      </div>
    );
  }
}

export default Form;
