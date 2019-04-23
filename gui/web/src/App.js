import React, { Component } from 'react';
import Header from './components/header/Header';
import './styles/styles.scss';
import './App.scss';

class App extends Component {
  render() {
    return (
      <div className="App">
        <Header version="1.04"/>
      </div>
    );
  }
}

export default App;
