import React, { Component } from 'react';
import styles from './ScreenHeader.module.scss';
import ScreenTitle from '../../atoms/ScreenTitle/ScreenTitle';
import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import Button from '../../atoms/Button/Button';


class ScreenHeader extends Component {

  render() {
    return (
      <div className={styles.wrapper}>
        <div  className={styles.titleWrapper}>
          { this.props.backButton && (
            <div className={styles.buttonWrapper}>
              <Link to="/">
                <Button
                icon="back"
                variant="transparent"
                hsize="round"/>
              </Link>
            </div>
          )}
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