import React, { Component } from 'react';
import styles from './Welcome.module.scss';
import classNames from 'classnames';
import Button from '../../atoms/Button/Button';
import symbol from '../../../assets/images/kelp-symbol.svg';

class Welcome extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpened: true,
    };
    this.close = this.close.bind(this);
  }

  open() {
    this.setState({
      isOpened: true,
    })
  }

  close() {
    this.setState({
      isOpened: false,
    })
  }

  render() {
    let isOpenedClass = this.state.isOpened ? styles.isOpened : null;

    let wrapperClasses = classNames(
      styles.wrapper,
      isOpenedClass,
    );

    return (
      <div className={wrapperClasses}>
        <div className={styles.window}>
          <Button 
            icon="close"
            size="small"
            variant="transparent"
            hsize="round"
            className={styles.closeButton} 
            onClick={this.close}
          />

          <div className={styles.image}>
            <img className={styles.symbol} src={symbol} alt="Kelp Symbol"/>
          </div>
          <div className={styles.content}>
            <h3 className={styles.title}>
              Welcome to Kelp 
              <span className={styles.version}>v1.04</span>
            </h3>

            <p className={styles.text}>
            Kelp is a free and open-source trading bot for the Stellar universal marketplace and centralized trading exchanges.</p>
            
            <p className={styles.text}>
            Kelp includes several configurable trading strategies and exchange integrations. You can use this GUI to define your own parameters to quickly get up and running with a trading bot in a matter of minutes.
            </p>

            <div className={styles.footer}>
              <Button variant="faded" onClick={this.close}>Start</Button>
            </div>
          </div>

        </div>
        <span className={styles.backdrop}/>
      </div>
    );
  }
}

export default Welcome;