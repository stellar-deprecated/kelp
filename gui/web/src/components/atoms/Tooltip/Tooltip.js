import React, { Component } from 'react';
import styles from './Tooltip.module.scss';

class Tooltip extends Component {
  render() {

    return (
      <div className={styles.wrapper}>
        <div className={styles.group}>
          <p className={styles.title}>Issuer:</p>
          <p className={styles.info}>interstellar.exchange</p>
        </div>
        <p className={styles.text}>Lorem ipsum dolor sit amet, consectetur adipiscing elit. In turpis sapien, pellentesque nec diam id, aliquet sodales ante.</p>
      </div>
    );
  }
}

export default Tooltip;