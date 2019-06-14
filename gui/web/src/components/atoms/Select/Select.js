import React, { Component } from 'react';
import styles from './Select.module.scss';

class Select extends Component {
  render() {
    let options = [];
    for (let i in this.props.options) {
      let o = this.props.options[i];
      options.push(<option key={o.value} value={o.value}>{o.text}</option>);
    }

    return (  
      <select className={styles.select} defaultValue={this.props.selected}>
        {options}
      </select>
    );
  }
}

export default Select;