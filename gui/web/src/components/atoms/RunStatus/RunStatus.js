import React, { Component } from 'react';
import styles from './RunStatus.module.scss';
import Icon from '../Icon/Icon';
import Constants from '../../../Constants';

class RunStatus extends Component {  
  render() {
    if (this.props.state === Constants.BotState.initializing) {
      return ( 
        <div className={styles.initializingWrapper}> 
          <Icon symbol="" width="8px" height="8px"/>
          <span className={styles.initializingLabel}>Initializing...</span>
        </div>
      );
    } else if (this.props.state === Constants.BotState.running) {
      return (
        <div className={styles.runningWrapper}>
          <div className={styles.runningIcon}>
            <Icon symbol="wave" width="22px" height="4px"/>
          </div>
          <span className={styles.runningLabel}>Running {this.props.timeRunning.toISOString().slice(14, -5)}</span>
        </div>
      );
    } else if (this.props.state === Constants.BotState.stopping) {
      return ( 
        <div className={styles.stoppingWrapper}> 
          <Icon symbol="" width="8px" height="8px"/>
          <span className={styles.stoppingLabel}>Stopping...</span>
        </div>
      );
    } else {
      return ( 
        <div className={styles.stoppedWrapper}> 
          <Icon symbol="stop" width="8px" height="8px"/>
          <span className={styles.stoppedLabel}>Stopped</span>
        </div>
      );
    }
  }
}

export default RunStatus;