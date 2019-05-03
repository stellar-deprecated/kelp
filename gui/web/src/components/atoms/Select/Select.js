import React, { Component } from 'react';
import styles from './Select.module.scss';

class Select extends Component {
  render() {
    return (
      <select className={styles.select}>
        <option value="sdex">SDEX</option>
        <option value="binance">Binance</option>
        <option value="other">Other</option>
      </select>
    );
  }
}

export default Select;