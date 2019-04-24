import React, { Component } from 'react';
import Header from './components/header/Header';
import Card from './components/card/Card';
import button from './components/button/Button.module.scss';
import style from './App.module.scss';
import grid from './styles/grid.module.scss';
import emptyIcon from './images/ico-empty.svg';
import { styles } from 'ansi-colors';

class App extends Component {
  render() {
    return (
      <div className="App">
        <Header version="v1.04"/>
        <div className={grid.container}>  
          {/* <div className={style.empty}>
            <img src={emptyIcon} className={style.icon} alt="Empty icon"/>
            <h2 className={style.title}>Your Kelp forest is empty</h2>
            <a href="#" className={button.medium}>Autogenerate your first test bot</a>
            <span className={style.separator}>or</span>
            <a href="#" className={button.link}>Create your first bot</a>
          </div> */}
          <div className={style.screenHeader}>
            <h1 className={style.screenTitle}>My Bots</h1>
            <button className={button.faded}>+ New Bot</button>
          </div>
          <Card />
        </div>
      </div>
    );
  }
}

export default App;
