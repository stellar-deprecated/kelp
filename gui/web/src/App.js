import React, { Component } from 'react';
import Header from './components/molecules/Header/Header';
import Card from './components/molecules/BotCard/BotCard';
import button from './components/atoms/Button/Button.module.scss';
import styles from './App.module.scss';
import grid from './components/_settings/grid.module.scss';
import emptyIcon from './assets/images/ico-empty.svg';

class App extends Component {
  render() {
    return (
      <div className="App">
        <Header version="v1.04"/>
        <div className={grid.container}>  
          <div className={styles.empty}>
            <img src={emptyIcon} className={styles.icon} alt="Empty icon"/>
            <h2 className={styles.title}>Your Kelp forest is empty</h2>
            <a href="#" className={button.medium}>Autogenerate your first test bot</a>
            <span className={styles.separator}>or</span>
            <a href="#" className={button.link}>Create your first bot</a>
          </div> 
          <div className={styles.screenHeader}>
            <h1 className={styles.screenTitle}>My Bots</h1>
            <button className={button.faded}>+ New Bot</button>
          </div>
          <Card />
        </div>
      </div>
    );
  }
}

export default App;
