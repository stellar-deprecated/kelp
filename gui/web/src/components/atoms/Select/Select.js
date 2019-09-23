import React, { Component } from 'react';
import styles from './Select.module.scss';
import Label from '../Label/Label';

class Select extends Component {
  render() {
    let selectedText = "";
    let options = [];
    for (let i in this.props.options) {
      let o = this.props.options[i];
      options.push(<option key={o.value} value={o.value}>{o.text}</option>);

      if (this.props.selected === o.value) {
        selectedText = o.text;
      }
    }

    if (this.props.readOnly) {
      return (<Label>{selectedText}</Label>);
    }

    return (  
      <select
        className={styles.select}
        value={this.props.selected}
        onChange={this.props.onChange}
        >
        {options}
      </select>
    );
  }
}

export default Select;