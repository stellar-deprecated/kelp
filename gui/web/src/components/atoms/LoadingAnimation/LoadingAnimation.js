import React, { Component } from 'react';
import styles from './LoadingAnimation.module.scss';


class LoadingAnimation extends Component {
  
  render() {
    return (
      <div className={styles.loader}></div>
    );
  }
}

export default LoadingAnimation;