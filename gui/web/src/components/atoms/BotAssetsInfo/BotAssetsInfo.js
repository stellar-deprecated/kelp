import React, { Component } from 'react';
import styles from './BotAssetsInfo.module.scss';
import InfoIcon from '../InfoIcon/InfoIcon';




class BotAssetsInfo extends Component {
  render() {
    return (
      <div>
        <div className={styles.baseAssetLine}>
          <span className={styles.assetCode}>XLM </span>
          <div className={styles.assetHelper}>
            {/* <InfoIcon/> */}
          </div>
          <span className={styles.assetValue}> 5,001.56</span>
        </div>
        <div className={styles.quoteAssetLine}>
          <span className={styles.assetCode}>USD </span>
          <div className={styles.assetHelper}>
            <InfoIcon/>
          </div> 
          <span className={styles.assetValue}> 5,001.56</span>
        </div>
      </div>
    )
  }
}

export default BotAssetsInfo;