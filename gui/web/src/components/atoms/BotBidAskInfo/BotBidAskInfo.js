import React, { Component } from 'react';
import styles from './BotBidAskInfo.module.scss';


class BotBidAskInfo extends Component {
  render() {
    return (
      <div>
        <div className={styles.spreadLine}>
          <span className={styles.spreadLabel}>Spread </span>
          <span className={styles.spreadValue}> $0.0014 (0.32%)</span>
        </div>
        <div className={styles.bidsLine}>
          <span className={styles.quoteNumber}>{this.props.num_bids}</span>
          <span className={styles.quoteNumber}> </span>
          <span className={styles.quotelabel}> bids</span>
        </div>
        <div className={styles.asksLine}>
          <span className={styles.quoteNumber}>{this.props.num_asks}</span>
          <span className={styles.quoteNumber}> </span>
          <span className={styles.quotelabel}> asks</span>
        </div>
      </div>
    )
  }
}

export default BotBidAskInfo;