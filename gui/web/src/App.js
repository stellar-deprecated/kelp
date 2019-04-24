import React, { Component } from 'react';
import Header from './components/header/Header';
import button from './components/button/Button.module.scss';
import style from './App.module.scss';
import emptyIcon from './images/ico-empty.svg';

class App extends Component {
  render() {
    return (
      <div className="App">
        <Header version="v1.04"/>
        <div className={style.empty}>
          <img src={emptyIcon} alt="Empty icon"/>
          <h2 className={style.title}>Your Kelp forest is empty</h2>
          <a href="#" className={button.medium}>Autogenerate your first test bot</a>
          <span className={style.separator}>or</span>
          <a href="#" className={button.link}>Create your first bot</a>
        </div> 
      </div>
    );
  }
}

export default App;
