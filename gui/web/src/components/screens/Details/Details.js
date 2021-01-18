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

  spacer() {
    return (
      <span className={styles.spacer}>
      </span>
    )
  }

  divider() {
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
          <ScreenHeader
            title="Harry the Green Plankton"
            backButtonFn={this.props.history.goBack}
            eventPrefix="details"
            >
              <PillGroup>
                <Pill number="1" type={'warning'}/>
                <Pill number="2" type={'error'}/>
              </PillGroup>
              <RunStatus />
              <StartStop/>
              <Button
                eventName={"details-close"}
                icon="options"
                size="large"
                variant="transparent"
                hsize="round"
                className={styles.optionsWrapper}
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
            
            {this.infoListItem('Trading platform', 'SDDAHRX2JB663N3OLKZSDDAHRX2JB663N3OLKZ')}
            {this.infoListItem('Trader account public key', '300 seconds')}
            {this.infoListItem('Base asset code', '0')}
            {this.infoListItem('Quote asset code', 'both')}
            {this.infoListItem('Quote asset issuer', 'interstellar.exchange')}

          </div>


          <div className={styles.advancedWrapper}>
            <div className={grid.container}>
              <h3 className={styles.advancedTitle}>Advanced trader settings</h3>

              {this.infoListItem('Source public key', 'SDDAHRX2JB663N3OLKZSDDAHRX2JB663N3OLKZ')}
              {this.infoListItem('Tick interval', '300 seconds')}
              {this.infoListItem('Randomized interval delay', '0')}
              {this.infoListItem('Submit mode', 'both')}
              {this.infoListItem('Delete cycles threshold', '0')}
              {this.infoListItem('Fill tracker', 'Off')}
              
              {this.divider()}

              {this.groupTitle('Fee')}
              {this.infoListItem('Fee capacity trigger', '0.5')}
              {this.infoListItem('Fee percentile computation', '90%')}
              {this.infoListItem('Maximum fee per operation', '5.000 stroops')}

              {this.divider()}

              {this.infoListItem('Decimal units for price', '6')}
              {this.infoListItem('Decimal units for volume', '1')}
              {this.infoListItem('Maximum volume of base units', '30')}
              {this.infoListItem('Maximum volume of quote units', '10')}

            </div>  
          </div>

          <div className={grid.container}>
            <SectionTitle className={styles.title}>
                Strategy Settings
            </SectionTitle>
            
            {this.infoListItem('Strategy type', 'buysell')}

            {this.divider()}

            {this.groupTitle('Price feed')}
            {this.infoListItem('Numerator price', 'Exchange > Kraken > XLM USD')}
            {this.infoListItem('Denominator price', 'Fixed > 1.0')}
            {this.infoListItem('Current Price', '0.0116 / 1 = 0.0116')}

            {this.divider()}

            {this.infoListItem('Amount of a base', '10')}
            {this.infoListItem('Spread of a market', '1%')}

          </div>

          <div className={styles.advancedWrapper}>
            <div className={grid.container}>
              <h3 className={styles.advancedTitle}>Advanced strategy settings</h3>
              
              {this.groupTitle('Levels')}
              {this.infoListItem('Spread', '2%')}
              {this.infoListItem('Amount', '10')}

              {this.spacer()}

              {this.infoListItem('Spread', '1%')}
              {this.infoListItem('Amount', '10')}
              
              {this.divider()}

              {this.infoListItem('Price tolerance', '1%')}
              {this.infoListItem('Amount tolerance', '0%')}
              {this.infoListItem('Rate offset percentage', '0%')}
              {this.infoListItem('Rate offset', '0')}
              {this.infoListItem('Use rate offset percent. first', 'On')}
            </div>  
          </div>
          
        </div>
      </div>
    )
  }
}

export default Details;
