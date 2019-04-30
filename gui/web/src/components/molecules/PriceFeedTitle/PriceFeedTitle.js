import React, { Component } from 'react';
import styles from './PriceFeedTitle.module.scss';
import classNames from 'classnames';
import ReloadButton from '../../atoms/ReloadButton/ReloadButton';
import Label from '../../atoms/Label/Label';

class PriceFeedTitle extends Component {
  static defaultProps = {
    groupTitle: "",
  }

  render() {
    return (
      <div className={styles.wrapper}>
        <Label>{this.props.label}</Label>
        <span className={styles.equals}>=</span>
        <span className={styles.value}>0,116</span>
        <ReloadButton/>
      </div>
    );
  }
}

export default PriceFeedTitle;