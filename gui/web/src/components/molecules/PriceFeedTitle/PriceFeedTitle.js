import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './PriceFeedTitle.module.scss';
import Label from '../../atoms/Label/Label';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';
import classNames from 'classnames';
import Button from '../../atoms/Button/Button';

class PriceFeedTitle extends Component {
  static propTypes = {
    label: PropTypes.string,
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

    return (
      <div className={styles.wrapper}>
        <Label>{this.props.label}</Label>
        <span className={styles.equals}>=</span>
        <div className={styles.valueWrapper}>
          <span className={valueClasses}>{this.props.price === null ? "<missing>" : this.props.price }</span>
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
          >
        </Button>
      </div>
    );
  }
}

export default PriceFeedTitle;