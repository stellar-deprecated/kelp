import React, { Component } from 'react';
import styles from './SegmentedControl.module.scss';

class SegmentedControl extends Component {
  render() {
    let segments = [];
    for (let i in this.props.segments) {
      let s = this.props.segments[i];

      let className = styles.segmentedControlItem;
      if (this.props.selected === s) {
        className = styles.segmentedControlItemSelected
      }

      segments.push(
        <li
          key={s}
          className={className}
          onClick={() => this.props.onSelect(s) }
          >
          {s}
        </li>
      );
    }

    return (
      <ul className={styles.segmentedControl}>
        {segments}
      </ul>
    );
  }
}

export default SegmentedControl;