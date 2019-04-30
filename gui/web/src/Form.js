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
import FieldGroup from './components/molecules/FieldGroup/FieldGroup';
import ScreenHeader from './components/molecules/ScreenHeader/ScreenHeader';
import AdvancedWrapper from './components/molecules/AdvacedWrapper/AdvancedWrapper';
import FormSection from './components/molecules/FormSection/FormSection';

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
              <FieldGroup>
                <Input size="large" value={'Harry the Green Plankton'}/>
              </FieldGroup>
              
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
              <FieldGroup>
                <Label>Trading Platform</Label>
                <Select/>
              </FieldGroup>
            </FormSection>
              
            <FormSection>
              <FieldGroup>
                <Label padding>Network</Label>
                <SegmentedControl/>
              </FieldGroup>
            </FormSection>
            
            <FormSection>
              <FieldGroup>
                <Label>Trader account secret key</Label>
                <Input error="Please enter a valid trader account secret key"/>
              </FieldGroup>
            </FormSection>

            <FormSection tip="Lorem ipsum dolor sit amet, consectetur adipiscing elit.  Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl.">
              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldGroup>
                    <Label>Base asset code</Label>
                    <Input/>
                  </FieldGroup>
                </div>

                <div className={grid.col8}>
                  <FieldGroup>
                    <Label>Base asset issuer</Label>
                    <Input/>
                  </FieldGroup>
                </div>
              </div>

              <div className={grid.row}>
                <div className={grid.col4}>
                  <FieldGroup>
                    <Label>Quote asset code</Label>
                    <Input/>
                  </FieldGroup>
                </div>

                <div className={grid.col8}>
                  <FieldGroup>
                    <Label>Quote asset issuer</Label>
                    <Input/>
                  </FieldGroup>
                </div>
              </div>
            </FormSection>
        </div>

        <AdvancedWrapper headerClass={grid.container}>
          <div className={grid.container}>
            <FormSection>
              <FieldGroup>
                <Label optional>Source account secret key</Label>
                <Input/>
              </FieldGroup>

              <FieldGroup>
                <Label>Tick interval</Label>
                <Input value="300" suffix="seconds"/>
              </FieldGroup>

              <FieldGroup>
                <Label>Randomized interval delay</Label>
                <Input value="0" suffix="miliseconds"/>
              </FieldGroup>

              <FieldGroup inline>
                <Switch></Switch>
                <Label>Maker only mode</Label>
              </FieldGroup>

              <FieldGroup>
                <Label>Delete cycles treshold</Label>
                <Input value="0"/>
              </FieldGroup>

              <FieldGroup inline>
                <Switch></Switch>
                <Label>Fill tracker</Label>
              </FieldGroup>

              <FieldGroup>
                <Label>Fill tracker duration</Label>
                <Input value="0" suffix="miliseconds"/>
              </FieldGroup>

              <FieldGroup>
                <Label>Fill tracker delete cycles threshold</Label>
                <Input value="0"/>
              </FieldGroup>


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
            
            <FieldGroup>
              <Label>Trading Platform</Label>
              <Input/>
            </FieldGroup>

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
