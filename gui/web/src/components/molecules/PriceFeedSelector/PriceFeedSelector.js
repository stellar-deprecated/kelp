import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './PriceFeedSelector.module.scss';
import Select from '../../atoms/Select/Select';
import Input from '../../atoms/Input/Input';
import grid from '../../../components/_styles/grid.module.scss';

class PriceFeedSelector extends Component {
  constructor(props) {
    super(props);

    this.renderComponentSingle = this.renderComponentSingle.bind(this);
    this.renderComponentRecursive = this.renderComponentRecursive.bind(this);
    this.getOptionItems = this.getOptionItems.bind(this);
    this.changeHandler = this.changeHandler.bind(this);
  }

  static propTypes = {
    optionsMetadata: PropTypes.object,
    values: PropTypes.arrayOf(PropTypes.string),
    onChange: PropTypes.func,
    readOnly: PropTypes.bool
  };

  getOptionItems(optionsObj) {
    let items = [];
    for (let key of Object.keys(optionsObj)) {
      items.push({
        value: optionsObj[key].value,
        text: optionsObj[key].text
      });
    }
    return items;
  }

  changeHandler(index, event) {
    let abridgedValues = this.props.values;
    abridgedValues[index] = event.target.value;
    // remove all elements after the index changed
    abridgedValues.splice(index + 1);

    let valuesToUpdate = [abridgedValues[0]];
    // set reasonable defaults
    let selectedOption = this.props.optionsMetadata.options[abridgedValues[0]];
    let i = 1;
    while (selectedOption.subtype !== null) {
      let curValue = "";
      if (selectedOption.subtype.type === "dropdown") {
        if (i < abridgedValues.length) {
          curValue = abridgedValues[i];
        } else {
          curValue = Object.keys(selectedOption.subtype.options)[0];
        }
        selectedOption = selectedOption.subtype.options[curValue];
      } else if (selectedOption.subtype.type === "text") {
        if (i < abridgedValues.length) {
          curValue = abridgedValues[i];
        } else {
          curValue = selectedOption.subtype.defaultValue;
        }
        selectedOption = selectedOption.subtype;
      }

      valuesToUpdate.push(curValue);
      i++;
    }

    this.props.onChange(valuesToUpdate);
  }

  renderComponentSingle(idx, metadata, value) {
    let className = grid.colPriceSelector;
    if (idx === 0) {
      className = grid.col5;
    }

    let selectedOption = {};
    let component = "";
    if (metadata.type === "text") {
      selectedOption = metadata;
      component = (
        <div className={className}>
          <Input
            value={value}
            type="string"
            onChange={(event) => this.changeHandler(idx, event)}
            readOnly={this.props.readOnly}
          />
        </div>
      );
    } else if (metadata.type === "dropdown") {
      let options = this.getOptionItems(metadata.options);
      selectedOption = metadata.options[value];
      component = (
        <div className={className}>
          <Select
            options={options}
            selected={value}
            onChange={(event) => this.changeHandler(idx, event)}
            readOnly={this.props.readOnly}
          />
        </div>
      );
    }

    return {
      selectedOption: selectedOption,
      componentFn: () => {
        return component;
      }
    };
  }

  renderComponentRecursive(idx, optionsMetadata, values) {
    let firstComponentWrapper = this.renderComponentSingle(idx, optionsMetadata, values[0]);
    let firstComponent = firstComponentWrapper.componentFn();
    let selectedOption = firstComponentWrapper.selectedOption;

    let secondComponent = "";
    if (values.length > 1) {
      let innerValues = [];
      for (let i = 1; i < values.length; i++) {
        innerValues.push(values[i]);
      }
      secondComponent = this.renderComponentRecursive(idx + 1, selectedOption.subtype, innerValues);
    }

    return (
      <div className={grid.row}>
        {firstComponent}
        {secondComponent}
      </div>
    );
  }

  render() {
    return (
      <div className={styles.wrapper}>
        {this.renderComponentRecursive(0, this.props.optionsMetadata, this.props.values)}
      </div>
    );
  }
}

export default PriceFeedSelector;