import React, { Component } from 'react';
import styles from './Badge.module.scss';




class Badge extends Component {
  render() {
    return (
        <div className={styles.error}>
          {this.props.number}
        </div>
    );
  }
}

export default Badge;