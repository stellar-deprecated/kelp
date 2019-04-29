import React, { Component } from 'react';
import styles from './RunStatus.module.scss';
import Icon from '../Icon/Icon';

class RunStatus extends Component {  
  render() {
    let statusComponent;

    if (this.props.timeRunning) {
      return ( 
        <div className={styles.runningWrapper}>
          <div className={styles.runningIcon}>
            <Icon symbol="wave"/>
          </div>
          <span className={styles.runningLabel}>Running {this.props.timeRunning.toISOString().slice(14, -5)}</span>
        </div>)
    } else {
      return ( 
        <div className={styles.stoppedWrapper}> 
          <Icon symbol="stop" width="7px" height="7px"/>
          <span className={styles.stoppedLabel}>Stopped</span>
        </div>
      )
    }
  }
}

export default RunStatus;