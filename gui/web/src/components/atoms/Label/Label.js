import React, { Component } from 'react';
import classNames from 'classnames';
import styles from './Label.module.scss';
import PropTypes from 'prop-types';

class Label extends Component {
  static propTypes = {
    padding: PropTypes.bool,
    optional: PropTypes.bool,
    className: PropTypes.string,
    disabled: PropTypes.bool
  };

  render() {
    const paddingClass = this.props.padding ? styles.padding : null;
    const disabledClass = this.props.disabled ? styles.disabled : null;
    const classNameList = classNames(
      styles.label,
      paddingClass,
      this.props.className,
      disabledClass,
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