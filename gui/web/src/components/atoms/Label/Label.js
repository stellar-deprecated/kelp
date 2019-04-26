import React, { Component } from 'react';
import styles from './Label.module.scss';

class Label extends Component {
  render() {
    return (
      <label className={styles.label}>{this.props.children}</label>
    );
  }
}

export default Label;