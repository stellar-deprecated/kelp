import React, { Component } from 'react';

import styles from './Details.module.scss';
import grid from '../../_styles/grid.module.scss';
import ScreenHeader from '../../molecules/ScreenHeader/ScreenHeader';
import StartStop from '../../atoms/StartStop/StartStop';
import RunStatus from '../../atoms/RunStatus/RunStatus';
import Button from '../../atoms/Button/Button';
import PillGroup from '../../molecules/PillGroup/PillGroup';
import Pill from '../../atoms/Pill/Pill';
import BotAssetsInfo from '../../atoms/BotAssetsInfo/BotAssetsInfo';
import BotExchangeInfo from '../../atoms/BotExchangeInfo/BotExchangeInfo';
import BotBidAskInfo from '../../atoms/BotBidAskInfo/BotBidAskInfo';

import chartImg from '../../../assets/images/chart-big.png';
import SectionTitle from '../../atoms/SectionTitle/SectionTitle';

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
            <PillGroup>
              <Pill number="1" type={'warning'}/>
              <Pill number="2" type={'error'}/>
            </PillGroup>
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
              
              <div className={styles.bidAskLine}>
                <BotBidAskInfo/>
              </div>
            </div>

          </div>

          <div className={styles.chartWrapper}>
            <img src={chartImg} alt="placeholder"/>
          </div>

          
        </div>
        <div className={styles.fullInfoWrapper}>
          <div className={grid.container}>
            <SectionTitle className={styles.title}>
              Trader Settings
            </SectionTitle>
            
            {this.infoListItem('Source Public key', 'SDDAHRX2JB663N3OLKZSDDAHRX2JB663N3OLKZ')}
            {this.infoListItem('Tick interval', '300 seconds')}
            {this.infoListItem('Randomized interval delay', '0')}
            {this.infoListItem('Submit mode', 'both')}
            {this.infoListItem('Delete cycles threshold', '0')}
            {this.infoListItem('Fill tracker', 'Off')}

          </div>


          <div className={styles.advancedWrapper}>
            <div className={grid.container}>
              <h3 className={styles.advancedTitle}>Advanced Settings</h3>

              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              
              {this.divider()}

              {this.groupTitle('Fee')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}

              {this.divider()}

              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}

            </div>  
          </div>

          <div className={grid.container}>
            <SectionTitle className={styles.title}>
                Trader Settings
            </SectionTitle>
            
            {this.infoListItem('Title', 'Information')}
            {this.infoListItem('Title', 'Information')}
            {this.infoListItem('Title', 'Information')}
            {this.infoListItem('Title', 'Information')}
            {this.infoListItem('Title', 'Information')}
          </div>

          <div className={styles.advancedWrapper}>
            <div className={grid.container}>
              <h3 className={styles.advancedTitle}>Advanced Settings</h3>
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
              {this.infoListItem('Title', 'Information')}
            </div>  
          </div>
          
        </div>
      </div>
    )
  }
}

export default Details;
