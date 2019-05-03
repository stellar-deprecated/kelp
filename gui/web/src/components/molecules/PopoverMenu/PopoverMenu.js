import React, { Component } from 'react';
import styles from './PopoverMenu.module.scss';

class PopoverMenu extends Component {
  render() {

    return (
      <div className={styles.wrapper}>
        <ul className={styles.list}>
          <li className={styles.item}>Edit</li>
          <li className={styles.item}>Copy</li>
          <li className={styles.itemDanger}>Delete</li>
        </ul>
      </div>
    );
  }
}

export default PopoverMenu;