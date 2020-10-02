import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Pill.module.scss';
import Icon from '../Icon/Icon';

class Pill extends Component {
  static propTypes = {
    type: PropTypes.string.isRequired,
    errors: PropTypes.array.isRequired,
    onClick: PropTypes.func.isRequired,
  };

  calculateErrorNumbers(errorObj) {
    const uniques = errorObj.length;
    const total = errorObj
        .map((errorElem) => errorElem.occurrences.length)
        .reduce((total, num) => (total + num), 0);
        
    return {
      uniques: uniques,
      total: total,
    };
  }

  render() {
    const errorNumbers = this.calculateErrorNumbers(this.props.errors);
    if (!errorNumbers.uniques) {
      return null;
    }

    let symbolName = "info";
    if (this.props.type === "warning") {
      symbolName = "warningSmall";
    } else if (this.props.type === "error") {
      symbolName = "errorSmall";
    }

    return (
      <div className={styles[this.props.type]} onClick={this.props.onClick}>
        <Icon className={styles.icon} symbol={symbolName} width={'11px'} height={'11px'}></Icon>
        <span>{errorNumbers.uniques}</span>
        <span className={styles.spacer}/>
        <span>(</span>
        <span>{errorNumbers.total}</span>
        <span>)</span>
      </div>
    );
  }
}

export default Pill;