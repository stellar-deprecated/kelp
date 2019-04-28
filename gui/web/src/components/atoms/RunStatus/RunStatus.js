import React, { Component } from 'react';
import styles from './RunStatus.module.scss';

class RunStatus extends Component {  
  render() {
    let statusComponent;

    if (this.props.timeRunning) {
      statusComponent = 
      <div>
        <i className={styles.runningIcon}></i>
        <span className={styles.runningLabel}>Running {this.props.timeRunning.toISOString().slice(11, -5)}</span>
      </div>;
    } else {
      statusComponent = 
      <div>
        <i className={styles.stoppedIcon}></i>
        <span className={styles.stoppedLabel}>Paused</span>
      </div>;
    }
    
    return (
      <div>
        {statusComponent}
      </div>
    );
  }
}

export default RunStatus;