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
      page: 1,
    };
    this.accept = this.accept.bind(this);
    this.setPage = this.setPage.bind(this);
    this.quit = this.quit.bind(this);
  }

  open() {
    this.setState({
      isOpened: true,
    })
  }

  accept() {
    this.setState({
      isOpened: false,
    })
  }

  setPage(pageNo) {
    this.setState({
      page: pageNo,
    })
  }

  quit() {
    this.props.quitFn()
  }

  render() {
    let isOpenedClass = this.state.isOpened ? styles.isOpened : null;

    let wrapperClasses = classNames(
      styles.wrapper,
      isOpenedClass,
    );

    const kelpLogo = (
      <div className={styles.image}>
        <img className={styles.symbol} src={symbol} alt="Kelp Symbol"/>
      </div>
    );

    const page1 = (
      <div className={styles.window}>
        {kelpLogo}
        <div className={styles.content}>
          <h3 className={styles.title}>
            Welcome to Kelp
          </h3>

          <p className={styles.text}>
          Kelp is a free and open-source trading bot for the Stellar Decentralized Exchange and centralized exchanges.
          </p>

          <p className={styles.text}>
          Kelp comes with multiple trading strategies out-of-the-box. This GUI of Kelp is limited to the Stellar Decentralized Exchange with the buysell strategy only.
          </p>

          <p className={styles.text}>
          You can use this GUI to define your own parameters to quickly get up and running with a trading bot in a matter of minutes.
          </p>

          <div className={styles.footer}>
            <Button variant="faded" onClick={this.setPage.bind(this, 2)}>Next</Button>
          </div>
        </div>
      </div>
    );
    const page2 = (
      <div className={styles.window}>
        {kelpLogo}
        <div className={styles.content}>
          <h3 className={styles.title}>
            Legal Disclaimer
          </h3>

          <p className={styles.text}>
          Prior to using this software, please note the following:
          </p>

          <ol className={styles.text}>
            <li>
              We do not recommend using this software on mainnet. This is experimental software, has many known bugs, and is not yet ready for use on mainnet. You could lose significant value by using this software on mainnet. 
            </li>
            <li>
              If you do alter the code and use it on mainnet, you acknowledge and agree that you fully assume full risk of doing so, and SDF shall not be held liable under any legal theory for loss of funds for any reason. 
            </li>
            <li>
              The experience you have with this software may be very different from what you will have on the final software that is made public for use on the mainnet.
            </li>
            <li>
              This software is provided under Apache 2.0. Please review the license carefully. 
            </li>
          </ol>

          <div className={styles.footer}>
            <Button variant="faded" onClick={this.accept}>Accept</Button>
            <Button variant="faded" onClick={this.quit}>Quit</Button>
          </div>
        </div>
      </div>
    );

    let pageDisplay = page1;
    if (this.state.page === 2) {
      pageDisplay = page2;
    }

    return (
      <div className={wrapperClasses}>
        {pageDisplay}
        <span className={styles.backdrop}/>
      </div>
    );
  }
}

export default Welcome;