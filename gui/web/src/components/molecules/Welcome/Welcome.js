import React, {Component} from 'react';
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
        <img className={styles.symbol} src={symbol} alt="Kelp Symbol" />
      </div>
    );

    const page1 = (
      <div className={styles.window}>
        {kelpLogo}
        <div className={styles.content}>
          <h3 className={styles.title}>
            Welcome to Kelp GUI (beta)
          </h3>

          <p className={styles.text}>
            Kelp GUI (beta) is a free and open-source trading bot for the Stellar Decentralized Exchange and centralized exchanges. Kelp is programmed to support multiple trading schemes. This GUI of Kelp is limited to the Stellar Decentralized Exchange with only the buysell scheme.
          </p>

          <p className={styles.text}>
            You can use this GUI to define your own parameters for each trading scheme to quickly get up and running with a trading bot in a matter of minutes.
          </p>

          <p className={styles.text}>
            Please note that SDF does not and cannot provide any recommendations or advice, express or implicit, concerning trading schemes, parameters, assets, or any other trading factor.
          </p>

          <div className={styles.footer}>
            <Button eventName="welcome-next" variant="faded" onClick={this.setPage.bind(this, 2)}>Next</Button>
          </div>
        </div>
      </div>
    );
    const page2 = (
      <div className={styles.window}>
        {kelpLogo}
        <div className={styles.content}>
          <h3 className={styles.title}>
            Kelp GUI (beta) Legal Disclaimer
          </h3>

          <p className={styles.text}>
            Prior to using this software, please note the following:
          </p>

          <ol className={styles.text}>
            <li>
              Please note that this is a beta version of Kelp GUI offered by the Stellar Development Foundation (“SDF”) which is still undergoing testing and debugging before its GA release. Known issues are being tracked on github <a href="https://github.com/stellar/kelp/issues" target="_blank">here</a>. The Kelp GUI (beta) software, and all content therein, are provided strictly on an “as is” and “as available” basis as part of this beta trial.
            </li>
            <li>
              You are responsible for determining the legal and regulatory requirements and restrictions concerning your use of this software. Do not use the software if your usage violates any laws or regulations applicable to you.
            </li>
            <li>
              SDF does not give any warranties, whether express or implied, as to the suitability, availability or usability of Kelp GUI (beta), it’s software or any of its content. Your use of Kelp GUI (beta) is solely at your own risk and SDF will not be liable for any loss, whether such loss is direct, indirect, special or consequential, suffered by any party as a result of their use of Kelp GUI (beta), its software or content.
            </li>
            <li>
              This software is provided under Apache 2.0. Please review the <a href="https://github.com/stellar/kelp/blob/master/LICENSE" target="_blank">license</a> carefully.
            </li>
            <li>
              Your use of Kelp GUI Beta is subject to the SDF <a href="https://stellar.org/terms-of-service" target="_blank">Terms</a> and <a href="https://stellar.org/privacy-policy" target="_blank">Privacy Policy</a>.
            </li>
          </ol>

          <div className={styles.footer}>
            <Button eventName="welcome-accept" variant="faded" onClick={this.accept}>Accept</Button>
            <Button eventName="welcome-quit" variant="faded" onClick={this.quit}>Quit</Button>
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
        <span className={styles.backdrop} />
      </div>
    );
  }
}

export default Welcome;