import React, { Component } from 'react';
import PropTypes from 'prop-types';

import styles from './EmptyList.module.scss';

import emptyIcon from '../../../assets/images/ico-empty.svg';
import Button from '../../atoms/Button/Button';


class EmptyList extends Component {
  static propTypes = {
    autogenerateFn: PropTypes.func.isRequired,
    createBotFn: PropTypes.func.isRequired,
  };

  render() {
    return (
      <div className={styles.empty}>
        <img src={emptyIcon} className={styles.icon} alt="Empty icon"/>
        <h2 className={styles.title}>Your Kelp forest is empty</h2>
        <Button eventName="main-newbot-autogen" onClick={this.props.autogenerateFn}>Autogenerate your first test bot on TestNet</Button>
        <span className={styles.separator}>or</span>
        <Button eventName="main-newbot-first" onClick={this.props.createBotFn} variant="link">Create your first bot</Button>
      </div> 
    );
  }
}

export default EmptyList;