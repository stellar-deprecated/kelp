import React, { Component } from 'react';
import styles from './PriceFeedFormula.module.scss';
import Label from '../../atoms/Label/Label';

class PriceFeedFormula extends Component {
  static defaultProps = {
    groupTitle: "",
  }

  render() {

    return (
      <div className={styles.wrapper}>
          <div className={styles.box}>
          <Label>Current price</Label>
          <div className={styles.formula}>0,116 / 1 = 0,116</div>
        </div>
      </div>
    );
  }
}

export default PriceFeedFormula;