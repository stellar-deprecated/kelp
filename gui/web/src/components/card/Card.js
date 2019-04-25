import React, { Component } from 'react';
import grid from '../../styles/grid.module.scss';
import button from '../Button/Button.module.scss';
import NotificationBadge from '../NotificationBadge/NotificationBadge';
import chartThumb from '../../assets/images/chart-thumb.png';
import infoIcon from '../../assets/images/ico-info.svg';
import stopIcon from '../../assets/images/ico-stop.svg';
import optionsIcon from '../../assets/images/ico-options.svg';
import style from './Card.module.scss';


class Card extends Component {
  render() {
    return (
      <div className={style.card}>
        <span className={style.status}></span>
        <button className={style.optionsMenu}><img src={optionsIcon}/></button>
        <div>
          <h2 className={style.title}>Sally te Blue Eel</h2>
          <div className={style.botDetailsLine}>
            <span className={style.netTag}>Test</span>
            <span className={style.exchange}>SDEX </span>
            <span className={style.lightText}> buysell</span>
          </div>
          <div>
            <div className={style.baseAssetLine}>
              <span className={style.textMono}>XLM </span>
              <i className={style.infoIcon}>
                <img src={infoIcon}/>
              </i>
              <span className={style.textMono}> 5,001.56</span>
            </div>
            <div className={style.quoteAssetLine}>
              <span className={style.textMono}>USD </span>
              <i className={style.infoIcon}>
                <img src={infoIcon}/>
              </i>
              <span className={style.textMono}> 5,001.56</span>
            </div>
          </div>
        </div>

        <div className={style.secondColumn}>
          <div className={style.notificationsLine}>
            <NotificationBadge number=" 2" type="warning"/>
          </div>
          <div className={style.spreadLine}>
            <span className={style.lightText}>Spread </span>
            <span className={style.textMono}> $0.0014 (0.32%)</span>
          </div>
          <div className={style.bidsLine}>
            <span className={style.textMono}>5 </span>
            <span className={style.textMono}> bids</span>
          </div>
          <div className={style.asksLine}>
            <span className={style.textMono}>3 </span>
            <span className={style.textMono}> asks</span>
          </div>
        </div>

        <div className={style.thirdColumn}>
          <img className={style.chartThumb} src={chartThumb} alt="chartThumb"/>
        </div>

        <div className={style.fourthColumn}>          
          <div className={style.statusDetails}>
            <i className={style.runningIcon}></i>
            <span className={style.runningTime}>Running 05:56</span>
          </div>
          <button className={style.startStopButton}>
            <img src={stopIcon}/>Stop
          </button>
        </div>

      </div>
    );
  }
}

export default Card;