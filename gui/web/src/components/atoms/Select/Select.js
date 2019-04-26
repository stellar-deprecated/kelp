import React, { Component } from 'react';
import './Select.css';

class Select extends Component {
  render() {
    return (
      <select>
        <option value="sdex">SDEX</option>
        <option value="binance">Binance</option>
        <option value="other">Other</option>
      </select>
    );
  }
}

export default Select;