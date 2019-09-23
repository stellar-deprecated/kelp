import React, { Component } from 'react';
import styles from './Switch.module.scss';

class Switch extends Component {
  render() {
    let onChange = this.props.onChange;
    if (this.props.readOnly) {
      onChange = () => {};
    }
    return (
      <div>
        <input
          className={styles.switch}
          type="checkbox"
          checked={this.props.value}
          onChange={onChange}
          />
      </div>
    );
  }
}

export default Switch;