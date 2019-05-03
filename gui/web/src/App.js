import React, { Component } from 'react';
import { BrowserRouter as Router, Route } from "react-router-dom";

import Header from './components/molecules/Header/Header';
import Bots from './components/screens/Bots/Bots';
import NewBot from './components/screens/NewBot/NewBot';
import Details from './components/screens/Details/Details';

class App extends Component {
  render() {
    return (
      <Router>
        <Header version="v1.04"/>
        <Route exact path="/" component={Bots} />
        <Route path="/new" component={NewBot} />
        <Route path="/details" component={Details} />
      </Router>
    );
  }
}

export default App;
