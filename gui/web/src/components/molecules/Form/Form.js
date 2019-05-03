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
    this.onClickSimulation = this.onClickSimulation.bind(this);
  }

  onClickSimulation() {
    this.setState({
      isLoading: true,
    })
  }

  render() {
    return (
      <div>
        <div className={grid.container}>
          <ScreenHeader title="New Bot" backButton>
            <Switch></Switch>
            <Label>Helper Fields</Label>
          </ScreenHeader>
            <FormSection>
              <Input size="large" value={'Harry the Green Plankton'}/>

              {/* Trader Settings */}
              <SectionTitle>
                Trader Settings
              </SectionTitle>
            </FormSection>
            
            <FormSection>
              <SectionDescription>
                Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
                Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl.
              </SectionDescription>
            </FormSection>

            <FormSection tip="Lorem ipsum dolor sit amet, consectetur adipiscing elit.  Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl.">
              <FieldItem>
                <Label>Trading Platform</Label>
                <Select/>
              </FieldItem>
            </FormSection>
              
            <FormSection>
              <FieldItem>
                <Label padding>Network</Label>
                <SegmentedControl/>
              </FieldItem>
            </FormSection>
            
            <FormSection>
              <FieldItem>
                <Label>Trader account secret key</Label>
                <Input error="Please enter a valid trader account secret key"/>
              </FieldItem>
            </FormSection>

            <FormSection tip="Lorem ipsum dolor sit amet, consectetur adipiscing elit.  Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl.">
              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldItem>
                    <Label>Base asset code</Label>
                    <Input/>
                  </FieldItem>
                </div>

                <div className={grid.col8}>
                  <FieldItem>
                    <Label>Base asset issuer</Label>
                    <Input/>
                  </FieldItem>
                </div>
              </div>

              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldItem>
                    <Label>Quote asset code</Label>
                    <Input/>
                  </FieldItem>
                </div>

                <div className={grid.col8}>
                  <FieldItem>
                    <Label>Quote asset issuer</Label>
                    <Input/>
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
                <Input/>
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
              onClick={this.onClickSimulation}>
              Create Bot
            </Button>
          </div>
        </div>
      
      </div>
    );
  }
}

export default Form;
