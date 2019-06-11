import React, { Component } from 'react';
import styles from './BotBidAskInfo.module.scss';


class BotBidAskInfo extends Component {
  render() {
    return (
      <div>
        <div className={styles.spreadLine}>
          <span className={styles.spreadLabel}>Spread </span>
          <span className={styles.spreadValue}> {this.props.spread_value + " (" + this.props.spread_pct + " %)"}</span>
        </div>
        <div className={styles.bidsLine}>
          <span className={styles.quoteNumber}>{this.props.num_bids < 0 ? "?" : this.props.num_bids}</span>
          <span className={styles.quoteNumber}> </span>
          <span className={styles.quotelabel}> bids</span>
        </div>
        <div className={styles.asksLine}>
          <span className={styles.quoteNumber}>{this.props.num_asks < 0 ? "?" : this.props.num_asks}</span>
          <span className={styles.quoteNumber}> </span>
          <span className={styles.quotelabel}> asks</span>
        </div>
      </div>
    )
  }
}

export default BotBidAskInfo;