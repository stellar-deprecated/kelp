import React, { Component } from 'react';
import styles from './FieldItem.module.scss';
import classNames from 'classnames';

class FieldItem extends Component {
  static defaultProps = {
    inline: false,
  }

  render() {
    var inlineClass = this.props.inline ? styles.inline : null;

    const wrapperClassList = classNames(
      styles.wrapper, 
      inlineClass,
    );

    return (
      <div className={wrapperClassList}>
        {this.props.children}
      </div>
    );
  }
}

export default FieldItem;