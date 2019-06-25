import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import styles from './Input.module.scss';

class Input extends Component {
  constructor(props) {
    super(props);

    this.checkType = this.checkType.bind(this);
    this.isString = this.isString.bind(this);
    this.isInt = this.isInt.bind(this);
    this.isFloat = this.isFloat.bind(this);
    this.isPercent = this.isPercent.bind(this);
    this.handleChange = this.handleChange.bind(this);
  }

  static defaultProps = {
    error: null,
    size: 'medium',
    disabled: false
  }

  static propTypes = {
    type: PropTypes.string.isRequired,       // types: string, int, float, percent
    value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
    error: PropTypes.string,
    size: PropTypes.string,
    disabled: PropTypes.bool,
    showError: PropTypes.bool
  };

  handleChange(event) {
    let checked = this.checkType(event.target.value);
    if (checked === null) {
      // if new input does not pass the check then don't allow an update
      return
    }

    let newEvent = event;
    if (this.props.type === "int" || this.props.type === "float") {
      // convert back to number so it is set correctly in the state
      newEvent = { target: { value: +checked } };
    } else if (this.props.type === "percent") {
      // convert back to representation passed in to complete the abstraction of a % value input
      newEvent = { target: { value: +checked / 100 } };
    }
    this.props.onChange(newEvent);
  }

  // returns "fixed-up" value as a string if type is correct, otherwise null
  checkType(input) {
    if (this.props.type === "string") {
      return this.isString(input);
    } else if (this.props.type === "int") {
      return this.isInt(input);
    } else if (this.props.type === "float") {
      return this.isFloat(input);
    } else if (this.props.type === "percent") {
      return this.isPercent(input);
    }
  }

  isString(input) {
    if (typeof input === "string") {
      return input;
    }
    return null;
  }

  // isInt always returns a string or null
  isInt(input) {
    if (isNaN(input)) {
      return null;
    }

    if (typeof input === "string") {
      if (input.includes(".")) {
        return null;
      }
      return input;
    }

    if (input % 1 === 0) {
      return input.toString();
    }
    return null;
  }

  // isFloat always returns a string or null
  isFloat(input) {
    if (isNaN(input)) {
      return null;
    }

    if (typeof input === "string") {
      if (input.includes(".")) {
        if (input.endsWith(".")) {
          return input + "0";
        }
        return input;
      }
      return input + ".00";
    }

    if (input % 1 === 0) {
      return input.toString() + ".00";
    }
    return input.toString();
  }

  // isPercent always returns a string or null
  isPercent(input) {
    let floatValue = this.isFloat(input);
    if (floatValue === null) {
      return null;
    }

    // convert to a number with the '+' before multiplying and then convert back to a string
    let pValue = +floatValue * 100
    let pValueString = pValue.toString();

    if (pValueString.includes(".")) {
      return pValueString;
    }
    return pValueString + ".00";
  }

  render() {
    const errorActive = this.props.showError ? styles.inputError : null;
    const inputClassList = classNames(
      styles.input, 
      styles[this.props.size],
      errorActive,
    );

    const suffixDisabled = this.props.disabled ? styles.disabled : null;
    const suffixClassList = classNames(
      styles.suffix, 
      suffixDisabled,
    );

    let value = this.checkType(this.props.value);
    if (value === null) {
      value = "invalid value: " + this.props.value + " for type=" + this.props.type;
    }

    return (
      <div className={styles.wrapper}>
        <input
          className={inputClassList}
          value={value}
          type="text"
          onChange={this.handleChange}
          disabled={this.props.disabled}
          />
        
        { this.props.suffix && (
        <p className={suffixClassList}>{this.props.suffix}</p>
        )}

        { this.props.showError && (
        <p className={styles.errorMessage}>{this.props.error}</p>
        )}
      </div>
    );
  }
}

export default Input;