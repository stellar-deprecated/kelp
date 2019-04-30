import React, { Component } from 'react';
import { BrowserRouter as Router, Route, Link } from "react-router-dom";

import Header from './components/molecules/Header/Header';
import Bots from './Bots';
import Form from './Form';
import Details from './Details';

class App extends Component {
  render() {
    return (
      <Router>
        <Header version="v1.04"/>
        <Route exact path="/" component={Bots} />
        <Route path="/form" component={Form} />
        <Route path="/details" component={Details} />
      </Router>
    );
  }
}

export default App;
