import React, { Component } from 'react';
import styles from './ScreenTitle.module.scss';

class ScreenTitle extends Component {

  render() {
    return (
      <h1 className={styles.title}>{this.props.children}</h1>
    );
  }
}

export default ScreenTitle;