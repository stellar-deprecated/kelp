import React, { Component } from 'react';
import styles from './BotExchangeInfo.module.scss';
import Badge from '../Badge/Badge';

class BotExchangeInfo extends Component {
  render() {
    return (
      <div className={styles.wrapper}>
        <Badge testnet={this.props.isTestnet}/>
        <span className={styles.exchange}>SDEX</span>
        <span className={styles.exchange}>  </span>
        <span className={styles.strategy}>{this.props.strategy}</span>
      </div>
    )
  }
}

export default BotExchangeInfo;