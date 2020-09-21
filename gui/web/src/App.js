import React, { Component } from 'react';
import { BrowserRouter as Router, Route } from "react-router-dom";
import Constants from './Constants';
import styles from './App.module.scss';
import Header from './components/molecules/Header/Header';
import Button from './components/atoms/Button/Button';
import Bots from './components/screens/Bots/Bots';
import NewBot from './components/screens/NewBot/NewBot';
import version from './kelp-ops-api/version';
import quit from './kelp-ops-api/quit';
import Welcome from './components/molecules/Welcome/Welcome';
// import Modal from './components/molecules/Modal/Modal';

let baseUrl = function () {
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
      version: "",
      kelp_errors: {},
    };

    this.setVersion = this.setVersion.bind(this);
    this.quit = this.quit.bind(this);
    this.addError = this.addError.bind(this);
    this.getErrors = this.getErrors.bind(this);
    this._asyncRequests = {};
  }

  componentDidMount() {
    this.setVersion()
  }

  setVersion() {
    var _this = this
    this._asyncRequests["version"] = version(baseUrl).then(resp => {
      if (!_this._asyncRequests["version"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["version"];
      if (!resp.includes("error")) {
        _this.setState({ version: resp });
      } else {
        setTimeout(_this.setVersion, 30000);
      }
    });
  }

  quit() {
    var _this = this
    this._asyncRequests["quit"] = quit(baseUrl).then(resp => {
      if (!_this._asyncRequests["quit"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }
      delete _this._asyncRequests["quit"];

      if (resp.status === 200) {
        window.close();
      }
    });
  }

  componentWillUnmount() {
    if (this._asyncRequests["version"]) {
      delete this._asyncRequests["version"];
    }
  }

  addError(backendError) {
    // TODO convert to hashID
    const ID = backendError.message

    // fetch object type from errors
    let kelp_errors = this.state.kelp_errors;

    if (!kelp_errors.hasOwnProperty(backendError.object_type)) {
      kelp_errors[backendError.object_type] = {};
    }
    let botErrors = kelp_errors[backendError.object_type];
    
    if (!botErrors.hasOwnProperty(backendError.object_name)) {
      botErrors[backendError.object_name] = {};
    }
    let namedError = botErrors[backendError.object_name];

    if (!namedError.hasOwnProperty(backendError.level)) {
      namedError[backendError.level] = {};
    }
    let levelErrors = namedError[backendError.level];

    if (!levelErrors.hasOwnProperty(ID)) {
      levelErrors[ID] = {
        occurrences: [],
        message: backendError.message,
      };
    }
    let idError = levelErrors[ID];

    // create new entry in list
    idError.occurrences.push(backendError.date);

    // trigger state change
    this.setState({ "kelp_errors": kelp_errors });
  }

  getErrors(object_type, object_name, level) {
    const kelp_errors = this.state.kelp_errors;

    if (!kelp_errors.hasOwnProperty(object_type)) {
      return null;
    }
    const botErrors = kelp_errors[object_type];
    
    if (!botErrors.hasOwnProperty(object_name)) {
      return null;
    }
    const namedError = botErrors[object_name];

    if (!namedError.hasOwnProperty(level)) {
      return null;
    }
    const levelErrors = namedError[level];
    return levelErrors;
  }

  render() {
    const enablePubnetBots = false;

    let banner = (<div className={styles.banner}>
      <Button
        className={styles.quit}
        size="small"
        onClick={this.quit}
      >
        Quit
      </Button>
      Kelp UI is only available on the Stellar Test Network
    </div>);

    const getBotErrors = this.getErrors.bind(this, Constants.ErrorType.bot);

    return (
      <div>
        <div>{banner}</div>
        <Router>
          <Header version={this.state.version}/>
          <Route exact path="/"
            render={(props) => <Bots {...props} baseUrl={baseUrl} addError={this.addError} getErrors={getBotErrors}/>}
            />
          <Route exact path="/new"
            render={(props) => <NewBot {...props} baseUrl={baseUrl} enablePubnetBots={enablePubnetBots}/>}
            />
          <Route exact path="/edit"
            render={(props) => <NewBot {...props} baseUrl={baseUrl} enablePubnetBots={enablePubnetBots}/>}
            />
          <Route exact path="/details"
            render={(props) => <NewBot {...props} baseUrl={baseUrl} enablePubnetBots={enablePubnetBots}/>}
            />
        </Router>
        {/* <Modal 
          type="error"
          title="Harry the Green Plankton has two warnings:"
          actionLabel="Go to bot settings"
          bullets={['Funds are low', 'Another warning example']}
        /> */}
        <Welcome quitFn={this.quit}/>
      </div>
    );
  }
}

export default App;
