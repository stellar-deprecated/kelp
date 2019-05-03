import React, { Component } from 'react';
import styles from './PriceFeedTitle.module.scss';
import Label from '../../atoms/Label/Label';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';
import classNames from 'classnames';
import Button from '../../atoms/Button/Button';


class PriceFeedTitle extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isLoading: false,
    };
    this.onClickSimulation = this.onClickSimulation.bind(this);
  }

  onClickSimulation() {
    this.startLoading()
    setTimeout(
      () => this.endLoading(),
      5000
    );
  }

  startLoading() {
    this.setState({
      isLoading: true,
    })
  }
  
  endLoading() {
    this.setState({
      isLoading: false,
    })
  }


  render() {
    const isLoading = this.state.isLoading ? styles.isLoading : null;

    const valueClasses = classNames(
      styles.value,
      isLoading,
    );

    return (
      <div className={styles.wrapper}>
        <Label>{this.props.label}</Label>
        <span className={styles.equals}>=</span>
        <div className ={styles.valueWrapper}>
            <span className={valueClasses}>0,116</span>
          { this.state.isLoading && (
            <div className={styles.loaderWrapper}>
              <LoadingAnimation/>
            </div>
          )}
        </div>
        
        <Button 
          onClick={this.onClickSimulation}
          icon="refresh"
          className={styles.button}
          variant="transparent"
          hsize="round"
          disabled={this.state.isLoading}
          >
        </Button>
      </div>
    );
  }
}

export default PriceFeedTitle;