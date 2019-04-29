import React, { Component } from 'react';
import styles from './Badge.module.scss';




class Badge extends Component {
  render() {

    return (
        <span className={this.props.test ? styles.test : styles.main}>
          {this.props.test ? 'Test' : 'Main'}
        </span>
    );
  }
}

export default Badge;