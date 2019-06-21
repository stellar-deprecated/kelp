import React, { Component } from 'react';
import styles from './PriceFeedFormula.module.scss';
import Label from '../../atoms/Label/Label';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';

class PriceFeedFormula extends Component {
  render() {
    let display = "";
    if (this.props.isLoading) {
      display = (
        <div className={styles.loaderWrapper}>
          <LoadingAnimation/>
        </div>
      );
    } else {
      let value = this.props.numerator/this.props.denominator;
      display = this.props.numerator + " / " + this.props.denominator + " = " + value;
    }

    return (
      <div className={styles.wrapper}>
          <div className={styles.box}>
          <Label>Current price</Label>
          <div className={styles.formula}>
            {display}
          </div>
        </div>
      </div>
    );
  }
}

export default PriceFeedFormula;