import React, { Component } from 'react';
import styles from './Input.module.scss';

class Input extends Component {

  render() {
    return (
      <input className={styles.input} type="text" name="fname"/>
    );
  }
}

export default Input;