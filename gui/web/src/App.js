import React, { Component } from 'react';
import { BrowserRouter as Router, Route, Link } from "react-router-dom";

import Header from './components/molecules/Header/Header';
import Bots from './Bots';
import Form from './Form';

class App extends Component {
  render() {
    return (
      <Router>
        <Header version="v1.04"/>
        <Route exact path="/" component={Bots} />
        <Route path="/form" component={Form} />
      </Router>
    );
  }
}

export default App;
