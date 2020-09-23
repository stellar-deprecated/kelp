import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Pill.module.scss';
import Icon from '../Icon/Icon';

class Pill extends Component {
  static propTypes = {
    type: PropTypes.string.isRequired,
    uniques: PropTypes.number,
    total: PropTypes.number,
  };

  render() {
    if (!this.props.uniques) {
      return null;
    }

    let symbolName = "info";
    if (this.props.type === "warning") {
      symbolName = "warningSmall";
    } else if (this.props.type === "error") {
      symbolName = "errorSmall";
    }

    return (
      <div className={styles[this.props.type]}>
        <Icon className={styles.icon} symbol={symbolName} width={'11px'} height={'11px'}></Icon>
        <span>{this.props.uniques}</span>
        <span className={styles.spacer}/>
        <span>(</span>
        <span>{this.props.total}</span>
        <span>)</span>
      </div>
    );
  }
}

export default Pill;