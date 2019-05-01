import React, { Component } from 'react';
import styles from './InfoIcon.module.scss';
import Icon from '../Icon/Icon';


class InfoIcon extends Component {
  render() {
    return (
      <div className={styles.wrapper}>
        <Icon className={styles.icon} symbol="info" width="8" height="8"/>
      </div>
    );
  }
}

export default InfoIcon;