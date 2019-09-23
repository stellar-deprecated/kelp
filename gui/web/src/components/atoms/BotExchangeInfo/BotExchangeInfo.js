import React, { Component } from 'react';
import styles from './BotExchangeInfo.module.scss';
import Badge from '../Badge/Badge';

class BotExchangeInfo extends Component {
  render() {
    let badgeValue = "pubnet";
    let badgeType = "main";
    if (this.props.isTestnet) {
      badgeValue = "testnet";
      badgeType = "test";
    }
    return (
      <div className={styles.wrapper}>
        <Badge value={badgeValue} type={badgeType}/>
        <span className={styles.exchange}>SDEX</span>
        <span className={styles.exchange}>  </span>
        <span className={styles.strategy}>{this.props.strategy}</span>
      </div>
    )
  }
}

export default BotExchangeInfo;