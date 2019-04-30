import React, { Component } from 'react';

import styles from './components/templates/Details/Details.module.scss';
import grid from './components/_settings/grid.module.scss';
import ScreenHeader from './components/molecules/ScreenHeader/ScreenHeader';
import Badge from './components/atoms/Badge/Badge';
import StartStop from './components/atoms/StartStop/StartStop';
import RunStatus from './components/atoms/RunStatus/RunStatus';
import Button from './components/atoms/Button/Button';
import PillGroup from './components/molecules/PillGroup/PillGroup';
import Pill from './components/atoms/Pill/Pill';
import BotBuySellInfo from './components/atoms/BotBuySellInfo/BotBuySellInfo';
import BotExchangeInfo from './components/atoms/BotExchangeInfo/BotExchangeInfo';
import BotAssetsInfo from './components/atoms/BotAssetsInfo/BotAssetsInfo';

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

          <div className={styles.mainInfo}>
            <div className={styles.firstInfoGroup}>
              <BotExchangeInfo/>
              <BotAssetsInfo/>
            </div>

            <div className={styles.secondInfoGroup}>
              <div className={styles.notificationsLine}>
                <PillGroup>
                  <Pill number="1" type={'warning'}/>
                  <Pill number="2" type={'error'}/>
                </PillGroup>
              </div>
              <BotBuySellInfo/>
            </div>
          </div>
        </div>
      </div>
    )
  }
}

export default Details;
