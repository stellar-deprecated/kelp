import React, { Component } from 'react';
import styles from './ScreenHeader.module.scss';
import ScreenTitle from '../../atoms/ScreenTitle/ScreenTitle';
import Button from '../../atoms/Button/Button';


class ScreenHeader extends Component {
  render() {
    let backButton = null;
    if (this.props.backButtonFn) {
      backButton = (
        <div className={styles.buttonWrapper}>
          <Button
          eventName={this.props.eventPrefix + "-back"}
          icon="back"
          variant="transparent"
          hsize="round"
          onClick={this.props.backButtonFn}/>
        </div>
      );
    }
    return (
      <div className={styles.wrapper}>
        <div  className={styles.titleWrapper}>
          {backButton}
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