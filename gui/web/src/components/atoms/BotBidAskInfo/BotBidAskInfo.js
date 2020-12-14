import React, { Component } from 'react';
import styles from './BotBidAskInfo.module.scss';
import functions from '../../../utils/functions';

class BotBidAskInfo extends Component {
  render() {
    let spreadValue = this.props.spread_value;
    if (spreadValue > 0) {
      spreadValue = functions.capSdexPrecision(spreadValue);
    }

    let spreadPct = this.props.spread_pct;
    if (spreadPct > 0) {
      spreadPct = spreadPct.toFixed(2);
    }

    return (
      <div>
        <div className={styles.spreadLine}>
          <span className={styles.spreadLabel}>Spread</span>
          <span className={styles.spreadValue}>{" " + spreadValue + " (" + spreadPct + " %)"}</span>
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