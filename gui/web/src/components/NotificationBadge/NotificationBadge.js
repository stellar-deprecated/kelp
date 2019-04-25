import React, { Component } from 'react';
import styles from './NotificationBadge.module.scss';
import warningIcon from '../../assets/images/ico-warning-small.svg';



class NotificationBadge extends Component {
  render() {
    return (
        <div className={styles.error}>
        <img src={warningIcon}/> 
        {this.props.number}
        </div>
    );
  }
}

export default NotificationBadge;