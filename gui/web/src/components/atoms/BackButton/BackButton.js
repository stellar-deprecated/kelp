import React, { Component } from 'react';

import styles from'./BackButton.module.scss';

import arrowBackIcon from '../../../assets/images/ico-arrow-back.svg';


class BackButton extends Component {
  render() {
    return (
      <button className={styles.button}>
        <img src={arrowBackIcon}/>
      </button>
    );
  }
}

export default BackButton;