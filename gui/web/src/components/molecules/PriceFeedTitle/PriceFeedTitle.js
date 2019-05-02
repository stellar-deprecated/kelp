import React, { Component } from 'react';
import styles from './PriceFeedTitle.module.scss';
import ReloadButton from '../../atoms/ReloadButton/ReloadButton';
import Label from '../../atoms/Label/Label';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';
import classNames from 'classnames';


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
    let isLoading = this.state.isLoading ? styles.isLoading : null;

    let valueClasses = classNames(
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
        
        <ReloadButton onClick={this.onClickSimulation}/>
      </div>
    );
  }
}

export default PriceFeedTitle;