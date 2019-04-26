import React, { Component } from 'react';
import styles from './RunStatus.module.scss';

class RunStatus extends Component {
  constructor(props) {
    super(props);
    this.state = {
      timeStarted: null,
      timeElapsed: null,
      isRunning: false,
    };

    this.toggleTimer = this.toggleTimer.bind(this);
  }

  toggleTimer() {
    if(this.state.isRunning){
      this.stopTimer();
    }
    else {
      this.startTimer();
    }
  }

  startTimer(){
    this.setState({
      timeStarted: new Date(),
      isRunning: true,
    });
    
    this.tick();
    this.timer = setInterval(
      () => this.tick(),
      1000
    );
  }

  stopTimer() {
    this.setState({
      timeStarted: null,
      isRunning: false,
    });
  }

  tick() {
    const timeNow = new Date();
    const diffTime = timeNow - this.state.timeStarted;
    
    const elapsed = new Date(diffTime);
    
    this.setState({
      timeElapsed: elapsed,
    });
  }
  
  render() {
    let statusComponent;

    if (this.state.isRunning) {
      statusComponent = 
      <div>
        <i className={styles.runningIcon}></i>
        <span className={styles.runningLabel}>Running {this.state.timeElapsed.toISOString().slice(11, -5)}</span>
      </div>;
    } else {
      statusComponent = 
      <div>
        <i className={styles.stoppedIcon}></i>
        <span className={styles.stoppedLabel}>Paused</span>
      </div>;
    }
    
    return (
      <div onClick={this.toggleTimer}>
        {statusComponent}
      </div>
    );
  }
}

//
export default RunStatus;