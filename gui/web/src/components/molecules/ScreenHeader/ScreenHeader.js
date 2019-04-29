import React, { Component } from 'react';
import styles from './ScreenHeader.module.scss';
import optionsIcon from '../../../assets/images/ico-options.svg';
import ScreenTitle from '../../atoms/ScreenTitle/ScreenTitle';
import BackButton from '../../atoms/BackButton/BackButton';
import { BrowserRouter as Router, Route, Link } from "react-router-dom";


class ScreenHeader extends Component {

  render() {
    return (
      <div className={styles.wrapper}>
        <div  className={styles.titleWrapper}>
          <div className={styles.buttonWrapper}>
            { this.props.backButton && (
              <Link to="/">
                <BackButton />
              </Link>
            )}
          </div>
          <ScreenTitle>{this.props.title}</ScreenTitle>
        </div>
        <div className={styles.childrenWrapper}>
          {this.props.children}
        </div>
      </div>
    );
  }
}

export default ScreenHeader;