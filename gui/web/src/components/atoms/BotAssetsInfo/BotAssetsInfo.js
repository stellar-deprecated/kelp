import React, { Component } from 'react';
import styles from './BotAssetsInfo.module.scss';
import InfoIcon from '../InfoIcon/InfoIcon';

class BotAssetsInfo extends Component {
  render() {
    const updatedDate = new Date(this.props.lastUpdated);
    const updatedDateLocalString = updatedDate.toLocaleString().replace(",", "");

    return (
      <div>
        <div className={styles.baseAssetLine}>
          <span className={styles.assetCode}>{this.props.assetBaseCode}</span>
          <span className={styles.assetCode}> </span>
          <div className={styles.assetHelper}>
            <InfoIcon issuer={this.props.assetBaseIssuer}/>
          </div>
          <span className={styles.assetValue}> </span>
          <span className={styles.assetValue}>{this.props.assetBaseBalance < 0 ? "?" : this.props.assetBaseBalance}</span>
        </div>
        <div className={styles.quoteAssetLine}>
          <span className={styles.assetCode}>{this.props.assetQuoteCode}</span>
          <span className={styles.assetCode}> </span>
          <div className={styles.assetHelper}>
            <InfoIcon issuer={this.props.assetQuoteIssuer}/>
          </div> 
          <span className={styles.assetValue}> </span>
          <span className={styles.assetValue}>{this.props.assetQuoteBalance < 0 ? "?" : this.props.assetQuoteBalance}</span>
        </div>
        <div className={styles.lastUpdatedLine}>
            <span className={styles.lastUpdatedField}>Updated</span>
            <span className={styles.lastUpdatedValue}>{updatedDateLocalString}</span>
          </div>
      </div>
    )
  }
}

export default BotAssetsInfo;