import React, { Component } from 'react';
import styles from './ScreenHeader.module.scss';
import ScreenTitle from '../../atoms/ScreenTitle/ScreenTitle';
import Button from '../../atoms/Button/Button';


class ScreenHeader extends Component {
  render() {
    return (
      <div className={styles.wrapper}>
        <div  className={styles.titleWrapper}>
          { this.props.backButton && (
            <div className={styles.buttonWrapper}>
              <Button
              icon="back"
              variant="transparent"
              hsize="round"/>
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