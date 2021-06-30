import React, { Component } from 'react';
import logo from '../../../assets/images/kelp-logo.svg';
import grid from '../../_styles/grid.module.scss';
import styles from './Header.module.scss';
import LogoutButton from '../../screens/LogAuth/LogoutButton';

class Header extends Component {
  render() {
    const auth0enabled = this.props.auth0enabled;
    return (
      <header className={styles.header}>
        <div className={grid.container}>
          <div className={styles.headerWrapper}>
            <div className={styles.logoWrapper}>
              <img src={logo} className={styles.logo} alt="Kelp logo" />
              <span className={styles.version}>{this.props.version}</span>
            </div>
            {auth0enabled ? (<LogoutButton className={styles.logoutButton}/>) : ""}
          </div>
        </div>
      </header>
    );
  }
}

export default Header;