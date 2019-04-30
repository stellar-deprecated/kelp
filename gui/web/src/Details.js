import React, { Component } from 'react';

import styles from './components/templates/Details/Details.module.scss';
import grid from './components/_settings/grid.module.scss';
import ScreenHeader from './components/molecules/ScreenHeader/ScreenHeader';
import Badge from './components/atoms/Badge/Badge';
import StartStop from './components/atoms/StartStop/StartStop';
import RunStatus from './components/atoms/RunStatus/RunStatus';
import Button from './components/atoms/Button/Button';
import Info from './components/atoms/Info/Info';
import PillGroup from './components/molecules/PillGroup/PillGroup';
import Pill from './components/atoms/Pill/Pill';

class Details extends Component {
  render() {
    return (
      <div>
        <div className={grid.container}>
          <ScreenHeader title="Harry the Green Plankton" backButton>
            <Badge test={true}/>
            <RunStatus />
            <StartStop/>
            <Button icon="options"/>
          </ScreenHeader>

          <div className={styleMedia.mainInfo}>
            <div className={styles.firstColumn}>
              <div className={styles.botDetailsLine}>
                <span className={styles.exchange}>SDEX</span>
                <span className={styles.lightText}>buysell</span>
              </div>
              <div>
                <div className={styles.baseAssetLine}>
                  <span className={styles.textMono}>XLM </span>
                  <span className={styles.textMono}> 5,001.56</span>
                </div>
                <div className={styles.quoteAssetLine}>
                  <span className={styles.textMono}>USD </span>
                  <Info/>
                  <span className={styles.textMono}> 5,001.56</span>
                </div>
              </div>
            </div>

            <div className={styles.secondColumn}>
              <div className={styles.notificationsLine}>
                <PillGroup>
                  <Pill number={this.props.warnings} type={'warning'}/>
                  <Pill number={this.props.errors} type={'error'}/>
                </PillGroup>
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


          </div>



          
        </div>
      </div>
    );
  }
}

export default Details;
