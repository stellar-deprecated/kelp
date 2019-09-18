import React, { Component } from 'react';
import styles from './PriceFeedFormula.module.scss';
import Label from '../../atoms/Label/Label';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';
import functions from '../../../utils/functions';

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
      const value = this.props.numerator/this.props.denominator;
      const baseDisplay = functions.assetDisplay(this.props.baseCode, this.props.baseIssuer);
      const quoteDisplay = functions.assetDisplay(this.props.quoteCode, this.props.quoteIssuer);

      display = (
        <div>
          <div>
            <span>{this.props.baseCode}</span>
            <span>{" / "}</span>
            <span>{this.props.quoteCode}</span>
            <span>{" = "}</span>
            <span>{this.props.numerator}</span>
            <span>{" / "}</span>
            <span>{this.props.denominator}</span>
            <span>{" = "}</span>
            <span className={styles.highlight}>{value}</span>
          </div>
          <div className={styles.explain}>
            <span>{"You are valuing 1 unit of "}</span>
            <span>{baseDisplay}</span>
            <span>{" at "}</span>
            <span className={styles.highlight}>{value}</span>
            <span>{" units of "}</span>
            <span>{quoteDisplay}</span>
          </div>
        </div>
      );
    }

    return (
      <div className={styles.wrapper}>
          <div className={styles.box}>
            <Label className={styles.heading}>{"Current Price"}</Label>
            <div className={styles.formula}>{display}</div>
        </div>
      </div>
    );
  }
}

export default PriceFeedFormula;