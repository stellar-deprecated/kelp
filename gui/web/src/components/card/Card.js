import React, { Component } from 'react';
import grid from '../../styles/grid.module.scss';
import button from '../button/Button.module.scss';
import chartThumb from '../../images/chart-thumb.png';
import style from './Card.module.scss';


class Card extends Component {
  render() {
    return (
      <div className={style.card}>
        <span className={style.status}></span>
        <div>
          <h2 className={style.title}>Sally te Blue Eel</h2>
          <div>
            <span className={style.netTag}>Test</span>
            <span className={style.exchange}>SDEX</span>
            <span className={style.lightText}>buysell</span>
          </div>
          <div>
            <div>
              <span className={style.textMono}>XLM</span>
              <span className={style.textMono}>5,001.56</span>
            </div>
            <div>
              <span className={style.textMono}>USD</span>
              <i className={style.infoIcon}></i>
              <span className={style.textMono}>5,001.56</span>
            </div>
          </div>
        </div>

        <div>
          <div>
            <span className={style.lightText}>Spead</span>
            <span className={style.textMono}>$0.0014 (0.32%)</span>
          </div>
          <div>
            <span className={style.textMono}>5</span>
            <span className={style.textMono}>bids</span>
          </div>
          <div>
            <span className={style.textMono}>3</span>
            <span className={style.textMono}>asks</span>
          </div>
        </div>

        <div>
          <img className={style.chartThumb} src={chartThumb} alt="chartThumb"/>
        </div>

        <div>
          <button>...</button>
          <div>
            <i></i>
            <span>Running 05:56</span>
          </div>
          <button><i></i>Stop</button>
        </div>

      </div>
    );
  }
}

export default Card;