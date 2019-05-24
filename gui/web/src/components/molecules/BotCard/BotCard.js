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
import Constants from '../../../Constants';

import start from '../../../kelp-ops-api/start';
import stop from '../../../kelp-ops-api/stop';
import getState from '../../../kelp-ops-api/getState';

class BotCard extends Component {
  constructor(props) {
    super(props);
    this.state = {
      timeStarted: null,
      timeElapsed: null,
      state: Constants.BotState.initializing,
    };
    this.toggleBot = this.toggleBot.bind(this);
    this.checkState = this.checkState.bind(this);
    this.startBot = this.startBot.bind(this);
    this.stopBot = this.stopBot.bind(this);
    this.tick = this.tick.bind(this);
    this._asyncRequests = {};
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
    showDetailsFn: PropTypes.func, 
    baseUrl: PropTypes.string, 
  };

  checkState() {
    if (this._asyncRequests["state"] == null) {
      var _this = this;
      this._asyncRequests["state"] = getState(this.props.baseUrl, this.props.name).then(resp => {
        _this._asyncRequests["state"] = null;
        let state = resp.trim();
        if (_this.state.state !== state) {
          _this.setState({
            state: state,
          });
        }
      });
    }
  }

  componentDidMount() {
    this.checkState();
    this._stateTimer = setInterval(this.checkState, 1000);
  }

  componentWillUnmount() {
    if (this._stateTimer) {
      clearTimeout(this._stateTimer);
      this._stateTimer = null;
    }

    if (this._tickTimer) {
      clearTimeout(this._tickTimer);
      this._tickTimer = null;
    }

    if (this._asyncRequests["state"]) {
      this._asyncRequests["state"].cancel();
      this._asyncRequests["state"] = null;
    }

    if (this._asyncRequests["start"]) {
      this._asyncRequests["start"].cancel();
      this._asyncRequests["start"] = null;
    }
    
    if (this._asyncRequests["stop"]) {
      this._asyncRequests["stop"].cancel();
      this._asyncRequests["stop"] = null;
    }
  }

  toggleBot() {
    if (this.state.state === Constants.BotState.running) {
      this.stopBot();
    } else {
      this.startBot();
    }
    this.checkState();
  }

  startBot() {
    var _this = this;
    this._asyncRequests["start"] = start(this.props.baseUrl, this.props.name).then(resp => {
      _this._asyncRequests["start"] = null;

      _this.setState({
        timeStarted: new Date(),
      }, () => {
        _this.checkState()
        _this.tick();
        _this._tickTimer = setInterval(_this.tick, 1000);
      });
    });
  }

  stopBot() {
    var _this = this;
    this._asyncRequests["stop"] = stop(this.props.baseUrl, this.props.name).then(resp => {
      _this._asyncRequests["stop"] = null;
      _this.setState({
        timeStarted: null,
      });
      clearTimeout(this._tickTimer);
      this._tickTimer = null;
    });
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
        <span className={this.state.state === Constants.BotState.running ? styles.statusRunning : styles.statusStopped}/>

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
          <h2 className={styles.title} onClick={this.props.showDetailsFn}>{this.props.name}</h2>
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
            state={this.state.state}
            timeRunning={this.state.timeElapsed}
          />
          <StartStop
            onClick={this.toggleBot} 
            state={this.state.state}
          />
        </div>

      </div>
    );
  }
}

export default BotCard;