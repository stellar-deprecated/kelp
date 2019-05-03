import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Pill from '../../atoms/Pill/Pill';
import RunStatus from '../../atoms/RunStatus/RunStatus';
import chartThumb from '../../../assets/images/chart-thumb.png';
import styles from './BotCard.module.scss';
import PillGroup from '../PillGroup/PillGroup';
import StartStop from '../../atoms/StartStop/StartStop';
import BotExchangeInfo from '../../atoms/BotExchangeInfo/BotExchangeInfo';
import BotAssetsInfo from '../../atoms/BotAssetsInfo/BotAssetsInfo';
import BotBidAskInfo from '../../atoms/BotBidAskInfo/BotBidAskInfo';
import Button from '../../atoms/Button/Button';


class BotCard extends Component {
  constructor(props) {
    super(props);
    this.state = {
      timeStarted: null,
      timeElapsed: null,
      isRunning: false,
    };
    this.toggleBot = this.toggleBot.bind(this);
  }

  static defaultProps = {
    name: '',
    test: true,
    warnings: 0,
    errors: 0, 
  }

  static propTypes = {
    name: PropTypes.string,
    test: PropTypes.bool,
    warnings: PropTypes.number,
    errors: PropTypes.number, 
  };

  toggleBot() {
    if(this.state.isRunning){
      this.stopBot();
    }
    else {
      this.startBot();
    }
  }

  startBot(){
    this.setState({
      timeStarted: new Date(),
      isRunning: true,
    }, () => {
      this.tick();
    
      this.timer = setInterval(
        () => this.tick(),
        1000
      );
    });
  }

  stopBot() {
    this.setState({
      timeStarted: null,
      timeElapsed: null,
      isRunning: false,
    });
    clearTimeout(this.timer);
  }

  tick() {
    let timeNow = new Date();
    let diffTime = timeNow - this.state.timeStarted;
    
    let elapsed = new Date(diffTime);
    
    this.setState({
      timeElapsed: elapsed,
    });
  }

  render() {
    return (
      <div className={styles.card}>
        <span className={this.state.isRunning ? styles.statusRunning : styles.statusStopped}></span>

        <Button
            icon="options"
            size="large"
            variant="transparent"
            hsize="round"
            className={styles.optionsMenu} 
            onClick={this.close}
        />

        <div className={styles.sortingArrows}>
          <Button
              icon="chevronUp"
              variant="transparent"
              hsize="round"
          />
          <Button
              icon="chevronDown"
              variant="transparent"
              hsize="round"
          />
        </div>

        <div className={styles.firstColumn}>
          <h2 className={styles.title}>{this.props.name}</h2>
          <div className={styles.botDetailsLine}>
            <BotExchangeInfo/>
          </div>
          <div>
            <BotAssetsInfo/>
          </div>
        </div>

        <div className={styles.secondColumn}>
          <div className={styles.notificationsLine}>
            <PillGroup>
              <Pill number={this.props.warnings} type={'warning'}/>
              <Pill number={this.props.errors} type={'error'}/>
            </PillGroup>
          </div>
          <BotBidAskInfo/>
        </div>

        <div className={styles.thirdColumn}>
          <img className={styles.chartThumb} src={chartThumb} alt="chartThumb"/>
        </div>

        <div className={styles.fourthColumn}>
          <RunStatus 
            className={styles.statusDetails} 
            timeRunning={this.state.timeElapsed}
          />
          <StartStop 
            onClick={this.toggleBot} 
            isRunning={this.state.isRunning}
          />
        </div>

      </div>
    );
  }
}

export default BotCard;