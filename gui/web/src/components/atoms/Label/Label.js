import React, { Component } from 'react';
import classNames from 'classnames';
import styles from './Label.module.scss';

class Label extends Component {

  render() {
    const padding = this.props.padding ? styles.padding : null;

    const classNameList = classNames(
      styles.label,
      padding
    );

    return (
      <label className={classNameList}>
      {this.props.children}
      {this.props.optional && (
        <span className={styles.optional}>Optional</span>
      )}
      </label>
    );
  }
}

export default Label;