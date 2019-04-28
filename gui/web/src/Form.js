import React, { Component } from 'react';

import Input from './components/atoms/Input/Input';
import Header from './components/molecules/Header/Header';

import styles from './components/templates/Form/Form.module.scss';
import button from './components/atoms/Button/Button.module.scss';
import grid from './components/_settings/grid.module.scss';

import emptyIcon from './assets/images/ico-empty.svg';
import Label from './components/atoms/Label/Label';
import SectionTitle from './components/atoms/SectionTitle/SectionTitle';
import SegmentedControl from './components/atoms/SegmentedControl/SegmentedControl';
import SectionDescription from './components/atoms/SectionDescription/SectionDescription';
import Button from './components/atoms/Button/Button';
import Select from './components/atoms/Select/Select';
import FieldGroup from './components/molecules/FieldGroup/FieldGroup';
import ScreenHeader from './components/molecules/ScreenHeader/ScreenHeader';
import Icon from './components/atoms/Icon/Icon';

class Form extends Component {
  render() {
    return (
      <div>
        <div className={grid.container}>
            <ScreenHeader title="New Bot" backButton={true}/>
            <div className={grid.row}>
              <div className={grid.col8}>
              
              <FieldGroup>
                <Input/>
              </FieldGroup>
              
              {/* Trader Settings */}
              <SectionTitle>
                Trader Settings
              </SectionTitle>
              
              <SectionDescription>
                Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
                Etiam purus nunc, rhoncus ac lorem eget, eleifend congue nisl.
              </SectionDescription>
              
              <FieldGroup>
                <Label>Name</Label>
                <Select/>
              </FieldGroup>
              
              <FieldGroup>
                <Label>Name</Label>
                <SegmentedControl/>
              </FieldGroup>
              
              <FieldGroup>
                <Label>Trader</Label>
                <Input/>
              </FieldGroup>
              
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
            </div>
          </div>
        </div>


        <div className={grid.container}>
          <Button size="large">Create Bot</Button>
        </div>
      </div>
    );
  }
}

export default Form;
