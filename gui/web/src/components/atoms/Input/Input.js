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
    disabled: PropTypes.bool
  };

  render() {
    const errorActive = this.props.error ? styles.inputError : null;

    const inputClassList = classNames(
      styles.input, 
      styles[this.props.size],
      errorActive,
    );

    return (
      <div className={styles.wrapper}>
        <input className={inputClassList} defaultValue={this.props.value} type="text"/>
        { this.props.suffix && (
        <p className={styles.suffix}>{this.props.suffix}</p>
        )}

        { this.props.error && (
        <p className={styles.errorMessage}>{this.props.error}</p>
        )}


      </div>
    );
  }
}

export default Input;