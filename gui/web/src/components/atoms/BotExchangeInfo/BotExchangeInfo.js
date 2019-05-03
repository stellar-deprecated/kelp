import React, { Component } from 'react';
import styles from './BotExchangeInfo.module.scss';
import Badge from '../Badge/Badge';



class BotExchangeInfo extends Component {
  render() {
    return (
      <div className={styles.wrapper}>
        <Badge test={true}/>
        <span className={styles.exchange}>SDEX </span>
        <span className={styles.strategy}> buysell</span>
      </div>
    )
  }
}

export default BotExchangeInfo;