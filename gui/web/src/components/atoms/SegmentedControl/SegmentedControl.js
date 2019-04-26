import React, { Component } from 'react';
import styles from './SegmentedControl.module.scss';

class SegmentedControl extends Component {

  render() {
    return (
      <ul className={styles.segmentedControl}>
        <li className={styles.segmentedControlItemSelected}>TestNet</li>
        <li className={styles.segmentedControlItem}>MainNet</li>
      </ul>
    );
  }
}

export default SegmentedControl;