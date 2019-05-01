import React, { Component } from 'react';
import logo from '../../../assets/images/kelp-logo.svg';
import grid from '../../_settings/grid.module.scss';
import button from '../../atoms/Button/Button.module.scss';
import styles from './Header.module.scss';
import Icon from '../../atoms/Icon/Icon';
import Button from '../../atoms/Button/Button';


class Header extends Component {
  render() {
    return (
      <header className={styles.header}>
        <div className={grid.container}>
          <div className={styles.headerWrapper}>
            <div className={styles.logoWrapper}>
              <img src={logo} className={styles.logo} alt="Kelp logo" />
              <span className={styles.version}>{this.props.version}</span>
            </div>
            <div className={styles.iconMenu}>
              <Button
                icon="day"
                size="large"
                variant="transparent"
                hsize="round"
              />
              <Button
                icon="help"
                size="large"
                variant="transparent"
                hsize="round"
              />
            </div>
          </div>
        </div>
      </header>
    );
  }
}

export default Header;