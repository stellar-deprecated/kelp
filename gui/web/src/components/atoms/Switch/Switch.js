import React, { Component } from 'react';
import styles from './Switch.module.scss';

class Switch extends Component {
  render() {
    return (
      <div>
        <input
          className={styles.switch}
          type="checkbox"
          checked={this.props.value}
          onChange={this.props.onChange}
          />
      </div>
    );
  }
}

export default Switch;