import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './PriceFeedDisplay.module.scss';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';
import classNames from 'classnames';
import Button from '../../atoms/Button/Button';
import functions from '../../../utils/functions';

class PriceFeedDisplay extends Component {
  static propTypes = {
    loading: PropTypes.bool,
    price: PropTypes.number,
    fetchPrice: PropTypes.func
  };

  render() {
    const isLoading = this.props.loading ? styles.isLoading : null;
    const valueClasses = classNames(
      styles.value,
      isLoading,
    );

    let priceCapped = this.props.price;
    if (priceCapped) {
      priceCapped = functions.capSdexPrecision(priceCapped);
    } else {
      priceCapped = "<missing>";
    }

    return (
      <div className={styles.wrapper}>
        <span className={styles.equals}>=</span>
        <div className={styles.valueWrapper}>
          <span className={valueClasses}>{priceCapped}</span>
          { this.props.loading && (
            <div className={styles.loaderWrapper}>
              <LoadingAnimation/>
            </div>
          )}
        </div>
        
        <Button 
          onClick={this.props.fetchPrice}
          icon="refresh"
          className={styles.button}
          variant="transparent"
          hsize="round"
          disabled={this.props.loading}
          />
      </div>
    );
  }
}

export default PriceFeedDisplay;