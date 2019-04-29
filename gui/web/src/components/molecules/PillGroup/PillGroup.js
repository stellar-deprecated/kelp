import React, { Component } from 'react';
import styles from './PillGroup.module.scss';


class PillGroup extends Component {

  render() {
    return (
      <div className={styles.groupss}>
        {this.props.children}
      </div>
    );
  }
}

export default PillGroup;