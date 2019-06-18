import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import styles from './Input.module.scss';

class Input extends Component {
  static defaultProps = {
    placeholder: null,
    value: null,
    eeror: null,
    size: 'medium',
    disabled: false    
  }

  static propTypes = {
    placebolder: PropTypes.string,
    value: PropTypes.string,
    error: PropTypes.string,
    size: PropTypes.string,
    disabled: PropTypes.bool,
    showError: PropTypes.bool
  };

  render() {
    const errorActive = this.props.showError ? styles.inputError : null;
    const strikethrough = this.props.strikethrough ? styles.strikethrough : null;

    const inputClassList = classNames(
      styles.input, 
      styles[this.props.size],
      errorActive,
      strikethrough,
    );

    return (
      <div className={styles.wrapper}>
        <input
          className={inputClassList}
          value={this.props.value}
          type="text"
          onChange={this.props.onChange}
          disabled={this.props.disabled}
          />
        { this.props.suffix && (
        <p className={styles.suffix}>{this.props.suffix}</p>
        )}

        { this.props.showError && (
        <p className={styles.errorMessage}>{this.props.error}</p>
        )}


      </div>
    );
  }
}

export default Input;