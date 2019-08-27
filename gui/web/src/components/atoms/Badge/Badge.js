import React, { Component } from 'react';
import styles from './Badge.module.scss';

class Badge extends Component {
  render() {
    return (
        <span className={this.props.testnet ? styles.test : styles.main}>
          {this.props.testnet ? 'Testnet' : 'Pubnet'}
        </span>
    );
  }
}

export default Badge;