import React, { Component } from 'react';
import styles from './FieldGroup.module.scss';

class FieldGroup extends Component {
  static defaultProps = {
    groupTitle: "",
  }

  render() {
    return (
      <div className={styles.wrapper}>
        {this.props.groupTitle && (
          <h4 className={styles.title}>{this.props.groupTitle}</h4>
        )}
        {this.props.children}
      </div>
    );
  }
}

export default FieldGroup;