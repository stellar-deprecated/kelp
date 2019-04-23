import React, { Component } from 'react';
import logo from '../../images/kelp-logo.svg';
import helpIcon from '../../images/ico-help.svg';
import dayIcon from '../../images/ico-day.svg';
import styles from './Header.module.scss';


class Header extends Component {
  render() {
    return (
      <header className={styles.header}>
        <div className={styles.wrapper}>
          <div className={styles.logoWrapper}>
            <img src={logo} className={styles.logo} alt="logo" />
            <i className={styles.version}>{this.props.version}</i>
          </div>
          <div className={styles.iconmenu}>
            <button>
              <img src={helpIcon}/>
            </button>
            <button>
              <img src={dayIcon}/>
            </button>
          </div>
        </div>
      </header>
    );
  }
}

export default Header;