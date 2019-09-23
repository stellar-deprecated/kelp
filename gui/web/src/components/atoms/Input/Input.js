import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import styles from './Input.module.scss';
import Label from '../Label/Label';

class Input extends Component {
  constructor(props) {
    super(props);

    this.checkValueError = this.checkValueError.bind(this);
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
    placeholder: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
    readOnly: PropTypes.bool,
    error: PropTypes.string,
    invokeChangeOnLoad: PropTypes.bool,
    triggerError: PropTypes.func,
    clearError: PropTypes.func,
    size: PropTypes.string,
    disabled: PropTypes.bool,
    onChange: PropTypes.func
  };

  componentWillMount() {
    this.checkValueError(this.props);
  }

  componentWillReceiveProps(nextProps) {
    if (this.props.value === nextProps.value) {
      return;
    }
    this.checkValueError(nextProps);
  }

  checkValueError(props) {
    let value = this.checkType(props.value);
    if (value === null) {
      value = props.value;
    }

    if (props.invokeChangeOnLoad) {
      this.handleChange({ target: { value: value } });
    }
  }

  handleChange(event) {
    let checked = this.checkType(event.target.value);
    if (checked === null) {
      this.props.onChange({ target: { value: event.target.value } });
      this.props.triggerError("invalid input value, expected type of input: " + this.props.type);
      return
    }

    let newEvent = event;
    if (this.props.type === "int" || this.props.type === "float") {
      // convert back to number so it is set correctly in the state
      newEvent = { target: { value: +checked } };
    } else if (this.props.type === "int_positive") {
      let val = +checked
      if (val <= 0) {
        this.props.onChange({ target: { value: val } });
        this.props.triggerError("invalid input value, needs to be a positive integer greater than 0");
        return
      }
      newEvent = { target: { value: val } };
    } else if (this.props.type === "int_nonnegative") {
      let val = +checked
      if (val < 0) {
        this.props.onChange({ target: { value: val } });
        this.props.triggerError("invalid input value, needs to be a non-negative integer >= 0");
        return
      }
      newEvent = { target: { value: val } };
    } else if (this.props.type === "float_positive") {
      let val = +checked
      if (val <= 0) {
        this.props.onChange({ target: { value: val } });
        this.props.triggerError("invalid input value, needs to be a positive decimal value greater than 0");
        return
      }
      newEvent = { target: { value: val } };
    } else if (this.props.type === "float_nonnegative") {
      let val = +checked
      if (val < 0) {
        this.props.onChange({ target: { value: val } });
        this.props.triggerError("invalid input value, needs to be a non-negative decimal value >= 0");
        return
      }
      newEvent = { target: { value: val } };
    } else if (this.props.type === "percent") {
      // convert back to representation passed in to complete the abstraction of a % value input
      // use event.target.value instead of checked here, because checked modified the value which is itself already modified
      newEvent = { target: { value: +event.target.value / 100 } };
    } else if (this.props.type === "percent_positive") {
      // use event.target.value instead of checked here, because checked modified the value which is itself already modified
      let val = +event.target.value
      if (val <= 0) {
        this.props.onChange({ target: { value: val / 100 } });
        this.props.triggerError("invalid input value, needs to be a positive percentage value greater than 0, represented as a decimal");
        return
      }
      // convert back to representation passed in to complete the abstraction of a % value input
      newEvent = { target: { value: val / 100 } };
    } else if (this.props.type === "percent_nonnegative") {
      // use event.target.value instead of checked here, because checked modified the value which is itself already modified
      let val = +event.target.value
      if (val < 0) {
        this.props.onChange({ target: { value: val } });
        this.props.triggerError("invalid input value, needs to be a non-negative percentage value >= 0, represented as a decimal");
        return
      }
      // convert back to representation passed in to complete the abstraction of a % value input
      newEvent = { target: { value: val / 100 } };
    }
    this.props.onChange(newEvent);
    if (this.props.type !== "string") {
      this.props.clearError();
    }
  }

  // returns "fixed-up" value as a string if type is correct, otherwise null
  checkType(input) {
    if (this.props.type === "string") {
      return this.isString(input);
    } else if (this.props.type === "int" || this.props.type === "int_positive" || this.props.type === "int_nonnegative") {
      return this.isInt(input);
    } else if (this.props.type === "float" || this.props.type === "float_positive" || this.props.type === "float_nonnegative") {
      return this.isFloat(input);
    } else if (this.props.type === "percent" || this.props.type === "percent_positive" || this.props.type === "percent_nonnegative") {
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
    if (isNaN(input) || input === "") {
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
    if (isNaN(input) || input === "") {
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
    const errorActive = this.props.error !== null ? styles.inputError : null;
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
      value = this.props.value;
    }

    let valueComponent = (
      <input
        className={inputClassList}
        defaultValue={value}
        placeholder={this.props.placeholder}
        type="text"
        onBlur={this.handleChange}
        disabled={this.props.disabled}
        readOnly={this.props.readOnly}
      />
    );
    if (this.props.readOnly) {
      valueComponent = (<Label disabled={this.props.disabled}>{value}</Label>);
    }

    return (
      <div>
        <div className={styles.wrapper} key={this.props.value}>
          {valueComponent}
          { this.props.suffix && (
          <p className={suffixClassList}>{this.props.suffix}</p>
          )}
        </div>
        { this.props.error !== null && (
        <div><p className={styles.errorMessage}>{this.props.error}</p></div>
        )}
      </div>
    );
  }
}

export default Input;