import React, { Component } from 'react';
import styles from './ScreenHeader.module.scss';
import optionsIcon from '../../../assets/images/ico-options.svg';
import ScreenTitle from '../../atoms/ScreenTitle/ScreenTitle';
import BackButton from '../../atoms/BackButton/BackButton';


class ScreenHeader extends Component {

  render() {
    return (
      <div className={styles.wrapper}>
        <div className={styles.buttonWrapper}>
          <BackButton />
        </div>
        <ScreenTitle>{this.props.title}</ScreenTitle>
      </div>
    );
  }
}

export default ScreenHeader;