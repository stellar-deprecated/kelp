import React, { Component } from 'react';
import styles from './Form.module.scss';
import grid from '../../_styles/grid.module.scss';
import Input from '../../atoms/Input/Input';
import Label from '../../atoms/Label/Label';
import SectionTitle from '../../atoms/SectionTitle/SectionTitle';
import Switch from '../../atoms/Switch/Switch';
import SegmentedControl from '../../atoms/SegmentedControl/SegmentedControl';
import SectionDescription from '../../atoms/SectionDescription/SectionDescription';
import Button from '../../atoms/Button/Button';
import Select from '../../atoms/Select/Select';
import FieldItem from '../FieldItem/FieldItem';
import ScreenHeader from '../ScreenHeader/ScreenHeader';
import AdvancedWrapper from '../AdvancedWrapper/AdvancedWrapper';
import FormSection from '../FormSection/FormSection';
import FieldGroup from '../FieldGroup/FieldGroup';
import PriceFeedSelector from '../PriceFeedSelector/PriceFeedSelector';
import PriceFeedTitle from '../PriceFeedTitle/PriceFeedTitle';
import PriceFeedFormula from '../PriceFeedFormula/PriceFeedFormula';
import RepeaterField from '../RepeaterField/RepeaterField';
import ErrorMessage from '../ErrorMessage/ErrorMessage';

class Form extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isLoading: false,
    };
    this.save = this.save.bind(this);
    this.collectConfigData = this.collectConfigData.bind(this);
  }

  collectConfigData() {
    // TODO collect config data from UI elements
    return this.props.configData;
  }

  save() {
    this.setState({
      isLoading: true,
    })
    let errorFields = this.props.saveFn(this.collectConfigData());
    if (errorFields) {
      // TODO mark errors
      return
    }

    this.props.router.goBack();
  }

  render() {
    let tradingPlatform = "sdex";
    if (this.props.configData.trader_config.trading_exchange && this.props.configData.trader_config.trading_exchange !== "") {
      tradingPlatform = this.props.configData.trader_config.trading_exchange;
    }

    let network = "TestNet";
    if (!this.props.configData.trader_config.horizon_url.includes("test")) {
      network = "PubNet";
    }

    let emptyStringIfXLM = (updatedCode) => {
      if (updatedCode === "XLM") {
        return "";
      }
      return null;
    };

    return (
      <div>
        <div className={grid.container}>
            <ScreenHeader title={this.props.title} backButtonFn={this.props.router.goBack}>
              <Switch/>
              <Label>Helper Fields</Label>
            </ScreenHeader>

            <FormSection>
              <Input
                size="large"
                value={this.props.configData.name}
                onChange={(event) => { this.props.onChange("name", event) }}
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

            <FormSection tip="Where do you want to trade: Stellar Decentralized Exchange (SDEX) or Kraken?">
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
            </FormSection>
              
            <FormSection>
              <FieldItem>
                <Label padding>Network</Label>
                <SegmentedControl
                  segments={[
                    "TestNet",
                    "PubNet",
                  ]}
                  selected={network}
                  />
              </FieldItem>
            </FormSection>
            
            <FormSection>
              <FieldItem>
                <Label>Trader account secret key</Label>
                <Input
                  value={this.props.configData.trader_config.trading_secret_seed}
                  onChange={(event) => { this.props.onChange("trader_config.trading_secret_seed", event) }}
                  error="Please enter a valid trader account secret key"
                  showError={false}
                  />
              </FieldItem>
            </FormSection>

            <FormSection tip="Lorem ipsum dolor sit amet, consectetur adipiscing elit.  Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl.">
              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldItem>
                    <Label>Base asset code</Label>
                    <Input
                      value={this.props.configData.trader_config.asset_code_a}
                      onChange={(event) => { this.props.onChange("trader_config.asset_code_a", event) }}
                      />
                  </FieldItem>
                </div>

                <div className={grid.col8}>
                  <FieldItem>
                    <Label>Base asset issuer</Label>
                    <Input
                      value={this.props.configData.trader_config.issuer_a}
                      onChange={(event) => { this.props.onChange("trader_config.issuer_a", event) }}
                      disabled={this.props.configData.trader_config.asset_code_a === "XLM"}
                      strikethrough={this.props.configData.trader_config.asset_code_a === "XLM"}
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
                      onChange={(event) => { this.props.onChange("trader_config.asset_code_b", event) }}
                      />
                  </FieldItem>
                </div>

                <div className={grid.col8}>
                  <FieldItem>
                    <Label>Quote asset issuer</Label>
                    <Input
                      value={this.props.configData.trader_config.issuer_b}
                      onChange={(event) => { this.props.onChange("trader_config.issuer_b", event) }}
                      disabled={this.props.configData.trader_config.asset_code_b === "XLM"}
                      strikethrough={this.props.configData.trader_config.asset_code_b === "XLM"}
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
                <Label optional>Source account secret key</Label>
                <Input
                  value={this.props.configData.trader_config.source_secret_seed}
                  onChange={(event) => { this.props.onChange("trader_config.source_secret_seed", event) }}
                  />
              </FieldItem>

              <FieldItem>
                <Label>Tick interval</Label>
                <Input value="300" suffix="seconds"/>
              </FieldItem>

              <FieldItem>
                <Label>Randomized interval delay</Label>
                <Input value="0" suffix="miliseconds"/>
              </FieldItem>

              <FieldItem inline>
                <Switch></Switch>
                <Label>Maker only mode</Label>
              </FieldItem>

              <FieldItem>
                <Label>Delete cycles treshold</Label>
                <Input value="0"/>
              </FieldItem>

              <FieldItem inline>
                <Switch></Switch>
                <Label>Fill tracker</Label>
              </FieldItem>

              <FieldItem>
                <Label>Fill tracker duration</Label>
                <Input value="0" suffix="miliseconds"/>
              </FieldItem>

              <FieldItem>
                <Label>Fill tracker delete cycles threshold</Label>
                <Input value="0"/>
              </FieldItem>

              <FieldGroup groupTitle="Fee">
                <div className={grid.row}>
                  <div className={grid.col5}>
                    <FieldItem>
                      <Label>Fee capacity trigger</Label>
                      <Input value="0.8"/>
                    </FieldItem>
                  </div>
                </div>
                <div className={grid.row}>
                  <div className={grid.col5}>
                    <FieldItem>
                      <Label>Fee capacity computation</Label>
                      <Input value="90" suffix="%"/>
                    </FieldItem>
                    </div>
                </div>
                <div className={grid.row}>
                  <div className={grid.col5}>
                    <FieldItem>
                      <Label>Maximum fee per operation</Label>
                      <Input value="5000" suffix="miliseconds"/>
                    </FieldItem>
                    </div>
                </div>
              </FieldGroup>

              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Decimal units for price</Label>
                    <Input value="6" />
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Decimal units for volume</Label>
                    <Input value="1" />
                  </FieldItem>
                </div>
              </div>

              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Min. volume of base units</Label>
                    <Input value="30.0" />
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Min. volume of quote units</Label>
                    <Input value="10.0" />
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
              Strategy Settings (buysell)
            </SectionTitle>
            
            <SectionDescription>
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
              Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl.
            </SectionDescription>
            

            <FieldGroup groupTitle="Price Feed">
              <FieldItem>
                <PriceFeedTitle label="Current numerator price"/>
                <PriceFeedSelector />
              </FieldItem>

              <FieldItem>
                <PriceFeedTitle label="Current denominator price"/>
                <PriceFeedSelector />
              </FieldItem>
              <PriceFeedFormula/>
            </FieldGroup>
            
            
            <div className={grid.row}>
              <div className={grid.col4}>
                <FieldItem>
                  <Label>Order size</Label>
                  <Input/>
                </FieldItem>
              </div>
            </div>
            
            <div className={grid.row}>
              <div className={grid.col4}>
                <FieldItem>
                  <Label>Spread of a market</Label>
                  <Input suffix="%"/>
                </FieldItem>
              </div>
            </div>
          </FormSection>
        </div>

        <AdvancedWrapper headerClass={grid.container}>
          <div className={grid.container}>
            <FormSection>
              <div className={grid.row}>
                <div className={grid.col8}>
                  <FieldGroup groupTitle="Levels">
                    <RepeaterField>
                          <FieldItem>
                            <Label>Spread</Label>
                            <Input suffix="%"/>
                          </FieldItem>
                          <FieldItem>
                            <Label>Amount</Label>
                            <Input/>
                          </FieldItem>
                    </RepeaterField>
                  </FieldGroup>
                </div>
              </div>
            </FormSection>

            <FormSection>
              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Price tolerance</Label>
                    <Input/>
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Amount tolerance</Label>
                    <Input/>
                  </FieldItem>
                </div>
              </div>

              <div className={grid.row}>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Rate offset percentage</Label>
                    <Input/>
                  </FieldItem>
                </div>
                <div className={grid.col5}>
                  <FieldItem>
                    <Label>Rate offset</Label>
                    <Input/>
                  </FieldItem>
                </div>
              </div>
            </FormSection>
          </div>
        </AdvancedWrapper>

        <div className={grid.container}>
          <ErrorMessage/>
          <div className={styles.formFooter}>
            <Button 
              icon="add" 
              size="large" 
              loading={this.state.isLoading} 
              onClick={this.save}>
              {this.props.saveText}
            </Button>
          </div>
        </div>
      
      </div>
    );
  }
}

export default Form;
