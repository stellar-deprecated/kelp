import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Badge.module.scss';

class Badge extends Component {
  static propTypes = {
    value: PropTypes.string.isRequired,
    type: PropTypes.string.isRequired,  // main, test, or message
  };

  render() {
    let clz = styles.message;
    if (this.props.type === "test") {
      clz = styles.test;
    } else if (this.props.type === "main") {
      clz = styles.main;
    }

    return (<span className={clz}>{this.props.value}</span>);
  }
}

export default Badge;