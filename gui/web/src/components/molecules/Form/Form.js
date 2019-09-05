import React, { Component } from 'react';
import styles from './Form.module.scss';
import grid from '../../_styles/grid.module.scss';
import Input from '../../atoms/Input/Input';
import Label from '../../atoms/Label/Label';
import SectionTitle from '../../atoms/SectionTitle/SectionTitle';
import Switch from '../../atoms/Switch/Switch';
// import SegmentedControl from '../../atoms/SegmentedControl/SegmentedControl';
import SectionDescription from '../../atoms/SectionDescription/SectionDescription';
import Button from '../../atoms/Button/Button';
// import Select from '../../atoms/Select/Select';
import FieldItem from '../FieldItem/FieldItem';
import ScreenHeader from '../ScreenHeader/ScreenHeader';
import AdvancedWrapper from '../AdvancedWrapper/AdvancedWrapper';
import FormSection from '../FormSection/FormSection';
import FieldGroup from '../FieldGroup/FieldGroup';
import PriceFeedAsset from '../PriceFeedAsset/PriceFeedAsset';
import PriceFeedFormula from '../PriceFeedFormula/PriceFeedFormula';
import Levels from '../Levels/Levels';
import ErrorMessage from '../ErrorMessage/ErrorMessage';
import newSecretKey from '../../../kelp-ops-api/newSecretKey';
import SecretKey from '../SecretKey/SecretKey';

class Form extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isSaving: false,
      isLoadingFormula: true,
      numerator: null,
      denominator: null,
    };
    this.setLoadingFormula = this.setLoadingFormula.bind(this);
    this.updateFormulaPrice = this.updateFormulaPrice.bind(this);
    this.save = this.save.bind(this);
    this.priceFeedAssetChangeHandler = this.priceFeedAssetChangeHandler.bind(this);
    this.updateLevel = this.updateLevel.bind(this);
    this.newLevel = this.newLevel.bind(this);
    this.hasNewLevel = this.hasNewLevel.bind(this);
    this.removeLevel = this.removeLevel.bind(this);
    this.newSecret = this.newSecret.bind(this);
    this.getError = this.getError.bind(this);
    this._emptyLevel = this._emptyLevel.bind(this);
    this._triggerUpdateLevels = this._triggerUpdateLevels.bind(this);
    this._fetchDotNotation = this._fetchDotNotation.bind(this);
    this._last_fill_tracker_sleep_millis = 1000;

    this._asyncRequests = {};
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
    if (ator === "numerator") {
      this.setState({
        isLoadingFormula: false,
        numerator: price
      })
    } else {
      this.setState({
        isLoadingFormula: false,
        denominator: price
      })
    }
  }

  _fetchDotNotation(obj, path) {
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
    if (!this.props.errorResp) {
      return null;
    }

    return this._fetchDotNotation(this.props.errorResp.fields, fieldKey);
  }

  save() {
    this.setState({
      isSaving: true,
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

    let mergeUpdateInstructions = {};
    mergeUpdateInstructions[feedUrlFieldName] = () => { return feedUrlValue };
    this.props.onChange(dataTypeFieldName, {target: {value: dataTypeValue}}, mergeUpdateInstructions);
  }

  updateLevel(levelIdx, fieldAmtSpread, value) {
    let newLevels = this.props.configData.strategy_config.levels;
    newLevels[levelIdx][fieldAmtSpread] = value;
    this._triggerUpdateLevels(newLevels);
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
    let newLevels = this.props.configData.strategy_config.levels;
    if (levelIdx >= newLevels.length) {
      return;
    }

    newLevels.splice(levelIdx, 1);
    if (newLevels.length === 0) {
      newLevels.push(this._emptyLevel());
    }

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

  render() {
    // let tradingPlatform = "sdex";
    // if (this.props.configData.trader_config.trading_exchange && this.props.configData.trader_config.trading_exchange !== "") {
    //   tradingPlatform = this.props.configData.trader_config.trading_exchange;
    // }

    let isTestNet = this.props.configData.trader_config.horizon_url.includes("test");
    // let network = "PubNet";
    // if (isTestNet) {
      // network = "TestNet";
    // }

    let error = "";
    if (this.props.errorResp) {
      error = (<ErrorMessage error={this.props.errorResp.error}/>);
    }

    let saveButtonDisabled = this.props.optionsMetadata == null;

    return (
      <div>
        <div className={grid.container}>
            <ScreenHeader title={this.props.title} backButtonFn={this.props.router.goBack}>
              {/* <Switch/>
              <Label>Helper Fields</Label> */}
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
                <Label padding>TestNet</Label>
                {/* <SegmentedControl
                  segments={[
                    "TestNet",
                    "PubNet",
                  ]}
                  selected={network}
                  onSelect={(selected) => {
                    // TODO use URI passed in from command line, or indicate to backend it's test/public
                    let newValue = "https://horizon-testnet.stellar.org";
                    if (selected === "PubNet") {
                      newValue = "https://horizon.stellar.org";
                    }
                    this.props.onChange("trader_config.horizon_url", {target: {value: newValue}});
                  }}
                  /> */}
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
                />
              </FieldItem>
            </FormSection>

            {/* <FormSection tip="Lorem ipsum dolor sit amet, consectetur adipiscing elit.  Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl."> */}
            <FormSection>
              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldItem>
                    <Label>Base asset code</Label>
                    <Input
                      value={this.props.configData.trader_config.asset_code_a}
                      type="string"
                      onChange={(event) => { this.props.onChange("trader_config.asset_code_a", event) }}
                      error={this.getError("trader_config.asset_code_a")}
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
                      onChange={(event) => { this.props.onChange("trader_config.asset_code_b", event) }}
                      error={this.getError("trader_config.asset_code_b")}
                      />
                  </FieldItem>
                </div>

                <div className={grid.col8}>
                  <FieldItem>
                    <Label>Quote asset issuer</Label>
                    <Input
                      value={this.props.configData.trader_config.issuer_b}
                      type="string"
                      onChange={(event) => { this.props.onChange("trader_config.issuer_b", event) }}
                      disabled={this.props.configData.trader_config.asset_code_b === "XLM"}
                      error={this.getError("trader_config.issuer_b")}
                      />
                  </FieldItem>
                </div>
              </div>
            </FormSection>
        </div>

        <AdvancedWrapper headerClass={grid.container}>
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
                  />
              </FieldItem>

              <FieldItem>
                <Label>Maximum Randomized Tick Interval Delay</Label>
                <Input
                  suffix="miliseconds"
                  value={this.props.configData.trader_config.max_tick_delay_millis}
                  type="int_positive"
                  onChange={(event) => { this.props.onChange("trader_config.max_tick_delay_millis", event) }}
                  error={this.getError("trader_config.max_tick_delay_millis")}
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
                  />
                <Label>Maker only mode</Label>  
              </FieldItem>

              <FieldItem>
                <Label>Delete cycles threshold</Label>
                <Input
                  value={this.props.configData.trader_config.delete_cycles_threshold}
                  type="int_positive"
                  onChange={(event) => { this.props.onChange("trader_config.delete_cycles_threshold", event) }}
                  error={this.getError("trader_config.delete_cycles_threshold")}
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
                  />
                <Label>Fill tracker</Label>
              </FieldItem>

              <FieldItem>
                <Label>Fill tracker duration</Label>
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
                          if (value.trim() === "0" || value.trim() === "") {
                            return this._last_fill_tracker_sleep_millis;
                          }

                          // just save the last value here and don't update the state in this function
                          this._last_fill_tracker_sleep_millis = value;
                          return null;
                        }
                      })
                    }
                  }
                  />
              </FieldItem>

              <FieldItem>
                <Label>Fill tracker delete cycles threshold</Label>
                <Input
                  value={this.props.configData.trader_config.fill_tracker_delete_cycles_threshold}
                  type="int_positive"
                  disabled={this.props.configData.trader_config.fill_tracker_sleep_millis === 0}
                  onChange={(event) => { this.props.onChange("trader_config.fill_tracker_delete_cycles_threshold", event) }}
                  error={this.getError("trader_config.fill_tracker_delete_cycles_threshold")}
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
                      />
                  </FieldItem>
                </div>
              </div>  

            </FormSection>
          </div>
        </AdvancedWrapper>


        {/* Stratefy Settings */}
        <div className={grid.container}>
          <FormSection>
            <SectionTitle>
              Strategy Settings <i>(buysell)</i>
            </SectionTitle>
            
            <SectionDescription>
              These settings refer to the strategy settings for the <i>buysell</i> strategy.
            </SectionDescription>
            

            <FieldGroup groupTitle="Price Feed">
              <FieldItem>
                <PriceFeedAsset
                  baseUrl={this.props.baseUrl}
                  onChange={(newValues) => this.priceFeedAssetChangeHandler("a", newValues)}
                  title="Current numerator price"
                  optionsMetadata={this.props.optionsMetadata}
                  type={this.props.configData.strategy_config.data_type_a}
                  feed_url={this.props.configData.strategy_config.data_feed_a_url}
                  onLoadingPrice={() => this.setLoadingFormula()}
                  onNewPrice={(newPrice) => this.updateFormulaPrice("numerator", newPrice)}
                  />
              </FieldItem>
              <FieldItem>
                <PriceFeedAsset
                  baseUrl={this.props.baseUrl}
                  onChange={(newValues) => this.priceFeedAssetChangeHandler("b", newValues)}
                  title="Current denominator price"
                  optionsMetadata={this.props.optionsMetadata}
                  type={this.props.configData.strategy_config.data_type_b}
                  feed_url={this.props.configData.strategy_config.data_feed_b_url}
                  onLoadingPrice={() => this.setLoadingFormula()}
                  onNewPrice={(newPrice) => this.updateFormulaPrice("denominator", newPrice)}
                  />
              </FieldItem>
              <PriceFeedFormula
                isLoading={this.state.isLoadingFormula || this.props.optionsMetadata == null}
                numerator={this.state.numerator}
                denominator={this.state.denominator}
                />
            </FieldGroup>
            
            <div className={grid.row}>
              <div className={grid.col8}>
                <FieldGroup groupTitle="Levels">
                  <Levels
                    levels={this.props.configData.strategy_config.levels}
                    updateLevel={(levelIdx, fieldAmtSpread, value) => { this.updateLevel(levelIdx, fieldAmtSpread, value) }}
                    newLevel={this.newLevel}
                    hasNewLevel={this.hasNewLevel}
                    onRemove={(levelIdx) => { this.removeLevel(levelIdx) }}
                    error={this.getError("strategy_config.levels")}
                    />
                </FieldGroup>
              </div>
            </div>
          </FormSection>
        </div>

        <AdvancedWrapper headerClass={grid.container}>
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
                  />
                <Label>Rate Offset Percent first</Label>
              </FieldItem>
            </FormSection>
          </div>
        </AdvancedWrapper>

        <div className={grid.container}>
          {error}
          <div className={styles.formFooter}>
            <Button 
              icon="add" 
              size="large" 
              loading={this.state.isSaving}
              disabled={saveButtonDisabled}
              onClick={this.save}
              >
              {this.props.saveText}
            </Button>
          </div>
        </div>
      
      </div>
    );
  }
}

export default Form;
