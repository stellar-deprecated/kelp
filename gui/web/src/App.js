import React, { Component } from 'react';
import { BrowserRouter as Router, Route } from "react-router-dom";

import Header from './components/molecules/Header/Header';
import Bots from './components/screens/Bots/Bots';
import NewBot from './components/screens/NewBot/NewBot';
import Details from './components/screens/Details/Details';
// import Welcome from './components/molecules/Welcome/Welcome';
// import Modal from './components/molecules/Modal/Modal';

import version from './kelp-ops-api/version';

let baseUrl = function() {
  let origin = window.location.origin
  if (process.env.REACT_APP_API_PORT) {
    let parts = origin.split(":")
    return parts[0] + ":" + parts[1] + ":" + process.env.REACT_APP_API_PORT;
  }
  return origin;
}()

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      version: ""
    };

    this._asyncRequests = {};
  }

  componentDidMount() {
    var _this = this
    this._asyncRequests["version"] = version(baseUrl).then(resp => {
      if (!_this._asyncRequests["version"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["version"];
      _this.setState({version: resp});
    });
  }

  componentWillUnmount() {
    if (this._asyncRequests["version"]) {
      delete this._asyncRequests["version"];
    }
  }

  render() {
    return (
      <div>
      <Router>
        <Header version={this.state.version}/>
        <Route exact path="/"
          render={(props) => <Bots {...props} baseUrl={baseUrl}/>}
          />
        <Route exact path="/new"
          render={(props) => <NewBot {...props} baseUrl={baseUrl}/>}
          />
        <Route exact path="/edit"
          render={(props) => <NewBot {...props} baseUrl={baseUrl}/>}
          />
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
