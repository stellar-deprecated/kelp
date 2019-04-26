import React, { Component } from 'react';
import styles from './FieldGroup.module.scss';

class FieldGroup extends Component {

  render() {
    return (
      <div className={styles.wrapper}>
        {this.props.children}
      </div>
    );
  }
}

export default FieldGroup;