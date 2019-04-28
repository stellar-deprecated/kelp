import React, { Component } from 'react';
import styles from './Pill.module.scss';
import warningIcon from '../../../assets/images/ico-warning-small.svg';



class Pill extends Component {
  render() {
    return (
        <div className={styles.error}>
          <img src={warningIcon}/> 
          {this.props.number}
        </div>
    );
  }
}

export default Pill;