  import React, { Component } from 'react';

import Input from './components/atoms/Input/Input';
import Header from './components/molecules/Header/Header';

import styles from './components/templates/Form/Form.module.scss';
import button from './components/atoms/Button/Button.module.scss';
import grid from './components/_settings/grid.module.scss';

import emptyIcon from './assets/images/ico-empty.svg';
import Label from './components/atoms/Label/Label';
import SectionTitle from './components/atoms/SectionTitle/SectionTitle';
import Switch from './components/atoms/Switch/Switch';
import SegmentedControl from './components/atoms/SegmentedControl/SegmentedControl';
import SectionDescription from './components/atoms/SectionDescription/SectionDescription';
import Button from './components/atoms/Button/Button';
import Select from './components/atoms/Select/Select';
import FieldItem from './components/molecules/FieldItem/FieldItem';
import ScreenHeader from './components/molecules/ScreenHeader/ScreenHeader';
import AdvancedWrapper from './components/molecules/AdvacedWrapper/AdvancedWrapper';
import FormSection from './components/molecules/FormSection/FormSection';
import FieldGroup from './components/molecules/FieldGroup/FieldGroup';
import PriceFeedSelector from './components/molecules/PriceFeedSelector/PriceFeedSelector';
import PriceFeedTitle from './components/molecules/PriceFeedTitle/PriceFeedTitle';
import PriceFeedFormula from './components/molecules/PriceFeedFormula/PriceFeedFormula';

class Form extends Component {
  render() {
    return (
      <div>
        <div className={grid.container}>
          <ScreenHeader title="New Bot" backButton={true}>
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
            
            
            <FieldItem>
              <Label>Trading Platform</Label>
              <Input/>
            </FieldItem>

          </FormSection>
        </div>


        <div className={grid.container}>
          <div className={styles.formFooter}>
            <Button size="large">Create Bot</Button>
          </div>
        </div>
      </div>
    );
  }
}

export default Form;
