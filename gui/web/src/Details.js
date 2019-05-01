import React, { Component } from 'react';

import styles from './components/templates/Details/Details.module.scss';
import grid from './components/_settings/grid.module.scss';
import ScreenHeader from './components/molecules/ScreenHeader/ScreenHeader';
import StartStop from './components/atoms/StartStop/StartStop';
import RunStatus from './components/atoms/RunStatus/RunStatus';
import Button from './components/atoms/Button/Button';
import PillGroup from './components/molecules/PillGroup/PillGroup';
import Pill from './components/atoms/Pill/Pill';
import BotAssetsInfo from './components/atoms/BotAssetsInfo/BotAssetsInfo';
import BotExchangeInfo from './components/atoms/BotExchangeInfo/BotExchangeInfo';
import BotBidAskInfo from './components/atoms/BotBidAskInfo/BotBidAskInfo';

import chartImg from './assets/images/chart-big.png';
import SectionTitle from './components/atoms/SectionTitle/SectionTitle';
import AdvancedWrapper from './components/molecules/AdvacedWrapper/AdvancedWrapper';

class Details extends Component {

  infoListItem(title, value) {
    return (
      <div className={styles.item}>
        <div className={grid.row}>
          <div className={grid.col3}>
            <div className={styles.label}>
              {title}
            </div>
          </div>
          <div className={grid.col6}>
            <div className={styles.value}>
              {value}
            </div>
          </div>
        </div>
      </div>
    )
  }

  divider(title, value) {
    return (
      <div className={styles.item}>
        <div className={grid.row}>
          <div className={grid.col3}>
          </div>
          <div className={grid.col6}>
            <span className={styles.divider}>
            </span>
          </div>
        </div>
      </div>
    )
  }

  groupTitle(title) {
    return (
      <div className={styles.item}>
        <div className={grid.row}>
          <div className={grid.col3}>
            <h4 className={styles.groupTitle}>
              {title}
            </h4>
          </div>
        </div>
      </div>
    )
  }

  render() {
    return (
      <div>
        <div className={grid.container}>
          <ScreenHeader title="Harry the Green Plankton" backButton>
            <RunStatus />
            <StartStop/>
            <Button
              icon="options"
              size="large"
              variant="transparent"
              hsize="round"
              className={styles.optionsMenu} 
              onClick={this.close}
            />
          </ScreenHeader>

          <div className={styles.mainInfo}>
            <div className={styles.firstInfoGroup}>
              <div className={styles.botDetailsLine}>
                <BotExchangeInfo/>
              </div>
              <BotAssetsInfo/>
            </div>
            <div className={styles.secondInfoGroup}>
              <div className={styles.notificationsLine}>
                <PillGroup>
                  <Pill number="1" type={'warning'}/>
                  <Pill number="2" type={'error'}/>
                </PillGroup>
              </div>
              <div className={styles.bidAskLine}>
                <BotBidAskInfo/>
              </div>
            </div>

          </div>

          <div className={styles.chartWrapper}>
            <img src={chartImg}/>
          </div>

          
        </div>
        <div className={styles.fullInfoWrapper}>
          <div className={grid.container}>
            <SectionTitle className={styles.title}>
              Trader Settings
            </SectionTitle>
            
            {this.infoListItem('trading Platform', 'World')}
            {this.infoListItem('Trader account public key', 'SDDAHRX2JB663N3OLKZSDDAHRX2JB663N3OLKZ')}
            {this.infoListItem('Hello', 'World')}
            {this.infoListItem('Hello', 'World')}
            {this.infoListItem('Hello', 'World')}

          </div>


          <div className={styles.advancedWrapper}>
            <div className={grid.container}>
              <h3 className={styles.advancedTitle}>Advanced Settings</h3>

              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              
              {this.divider()}

              {this.groupTitle('Fee')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}

              {this.divider()}

              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}

            </div>  
          </div>

          <div className={grid.container}>
            <SectionTitle className={styles.title}>
                Trader Settings
            </SectionTitle>
            
            {this.infoListItem('Hello', 'World')}
            {this.infoListItem('Hello', 'World')}
            {this.infoListItem('Hello', 'World')}
            {this.infoListItem('Hello', 'World')}
            {this.infoListItem('Hello', 'World')}
          </div>

          <div className={styles.advancedWrapper}>
            <div className={grid.container}>
              <h3 className={styles.advancedTitle}>Advanced Settings</h3>
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
              {this.infoListItem('Hello', 'World')}
            </div>  
          </div>
          
        </div>
      </div>
    )
  }
}

export default Details;
