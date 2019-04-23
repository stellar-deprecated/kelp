import React, { Component } from 'react';
import styles from './Button.module.scss';



class Button extends Component {
  render() {
    return (
        <button className={styles.error}>Error Button</button>
        // <h1>Hello {this.props.name}</h1>
    );
  }
}

export default Button;