import React, { Component } from 'react';
import logo from '../../images/kelp-logo.svg';
import styles from './Header.module.scss';


class Header extends Component {
  render() {
    return (
        <header className={styles.header}>
            <div className={styles.logoWrapper}>
                <img src={logo} className={styles.logo} alt="logo" />
                <i className="version">{this.props.version}</i>
            </div>
        </header>
    );
  }
}

export default Header;