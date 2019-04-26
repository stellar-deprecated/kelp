import React, { Component } from 'react';

import NotificationBadge from '../../atoms/NotificationBadge/NotificationBadge';
import RunStatus from '../../atoms/RunStatus/RunStatus';

import chartThumb from '../../../assets/images/chart-thumb.png';
import infoIcon from '../../../assets/images/ico-info.svg';
import stopIcon from '../../../assets/images/ico-stop.svg';
import optionsIcon from '../../../assets/images/ico-options.svg';

import button from '../../atoms/Button/Button.module.scss';
import grid from '../../_settings/grid.module.scss';
import styles from './BotCard.module.scss';


class BotCard extends Component {
  render() {
    return (
      <div className={styles.card}>
        <span className={styles.status}></span>
        <button className={styles.optionsMenu}><img src={optionsIcon}/></button>
        <div>
          <h2 className={styles.title}>Sally te Blue Eel</h2>
          <div className={styles.botDetailsLine}>
            <span className={styles.netTag}>Test</span>
            <span className={styles.exchange}>SDEX </span>
            <span className={styles.lightText}> buysell</span>
          </div>
          <div>
            <div className={styles.baseAssetLine}>
              <span className={styles.textMono}>XLM </span>
              <i className={styles.infoIcon}>
                <img src={infoIcon}/>
              </i>
              <span className={styles.textMono}> 5,001.56</span>
            </div>
            <div className={styles.quoteAssetLine}>
              <span className={styles.textMono}>USD </span>
              <i className={styles.infoIcon}>
                <img src={infoIcon}/>
              </i>
              <span className={styles.textMono}> 5,001.56</span>
            </div>
          </div>
        </div>

        <div className={styles.secondColumn}>
          <div className={styles.notificationsLine}>
            <NotificationBadge number=" 2" type="warning"/>
          </div>
          <div className={styles.spreadLine}>
            <span className={styles.lightText}>Spread </span>
            <span className={styles.textMono}> $0.0014 (0.32%)</span>
          </div>
          <div className={styles.bidsLine}>
            <span className={styles.textMono}>5 </span>
            <span className={styles.textMono}> bids</span>
          </div>
          <div className={styles.asksLine}>
            <span className={styles.textMono}>3 </span>
            <span className={styles.textMono}> asks</span>
          </div>
        </div>

        <div className={styles.thirdColumn}>
          <img className={styles.chartThumb} src={chartThumb} alt="chartThumb"/>
        </div>

        <div className={styles.fourthColumn}>
          <RunStatus className={styles.statusDetails}/>

          <button className={styles.startStopButton}>
            <img src={stopIcon}/>Stop
          </button>
        </div>

      </div>
    );
  }
}

export default BotCard;