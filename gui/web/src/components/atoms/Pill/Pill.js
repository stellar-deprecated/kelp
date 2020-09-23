import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Pill.module.scss';
import Icon from '../Icon/Icon';

class Pill extends Component {
  static propTypes = {
    type: PropTypes.string.isRequired,
    number: PropTypes.number,
  };

  render() {
    if (!this.props.number) {
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
        <span>{this.props.number}</span>
      </div>
    );
  }
}

export default Pill;