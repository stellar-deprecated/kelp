import React, { Component } from 'react';
import styles from './PopoverMenu.module.scss';

class PopoverMenu extends Component {
  render() {

    return (
      <div className={styles.wrapper}>
        <div className={styles.list}>
          <div className={styles.item}>Edit</div>
          <div className={styles.item}>Copy</div>
          <div className={styles.itemDanger}>Delete</div>
        </div>
      </div>
    );
  }
}

export default PopoverMenu;