import React, { Component } from 'react';

import styles from './EmptyList.module.scss';

import emptyIcon from '../../../assets/images/ico-empty.svg';
import Button from '../../atoms/Button/Button';


class EmptyList extends Component {
  render() {
    return (
      <div className={styles.empty}>
        <img src={emptyIcon} className={styles.icon} alt="Empty icon"/>
        <h2 className={styles.title}>Your Kelp forest is empty</h2>
        <Button onClick={this.props.onClick}>Autogenerate your first test bot</Button>
        <span className={styles.separator}>or</span>
        <Button onClick={this.props.onClick} variant="link">Create your first bot</Button>
      </div> 
    );
  }
}

export default EmptyList;