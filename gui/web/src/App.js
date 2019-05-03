import React, { Component } from 'react';
import { BrowserRouter as Router, Route } from "react-router-dom";

import Header from './components/molecules/Header/Header';
import Bots from './components/screens/Bots/Bots';
import NewBot from './components/screens/NewBot/NewBot';
import Details from './components/screens/Details/Details';
import Welcome from './components/molecules/Welcome/Welcome';
import Modal from './components/molecules/Modal/Modal';

class App extends Component {
  render() {
    return (
      <div>
      <Router>
        <Header version="v1.04"/>
        <Route exact path="/" component={Bots} />
        <Route path="/new" component={NewBot} />
        <Route path="/details" component={Details} />
      </Router>
      {/* <Modal 
        type="error"
        title="Harry the Green Plankton has two warnings:"
        actionLabel="Go to bot settings"
        bullets={['Funds are low', 'Another warning example']}
      />
      <Welcome/> */}
      </div>
    );
  }
}

export default App;
