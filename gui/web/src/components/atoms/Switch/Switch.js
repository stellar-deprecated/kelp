import React, { Component } from 'react';
import styles from './Switch.module.scss';

class Switch extends Component {
  render() {
    return (
      <div>
        <input className={styles.switch} type="checkbox" value={this.props.value}/>
      </div>
    );
  }
}

export default Switch;