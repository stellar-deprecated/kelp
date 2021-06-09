import React, { Component, useEffect } from 'react';
import { BrowserRouter as Router, Route } from "react-router-dom";
import Constants from './Constants';
import styles from './App.module.scss';
import Header from './components/molecules/Header/Header';
import Button from './components/atoms/Button/Button';
import Bots from './components/screens/Bots/Bots';
import NewBot from './components/screens/NewBot/NewBot';
import version from './kelp-ops-api/version';
import serverMetadata from './kelp-ops-api/serverMetadata';
import quit from './kelp-ops-api/quit';
import fetchKelpErrors from './kelp-ops-api/fetchKelpErrors';
import removeKelpErrors from './kelp-ops-api/removeKelpErrors';
import Welcome from './components/molecules/Welcome/Welcome';
import LoginRedirect  from './components/screens/LogAuth/LoginRedirect';
import { interceptor } from './kelp-ops-api/interceptor';
import { withAuth0 } from '@auth0/auth0-react';
import { withAuthenticationRequired } from '@auth0/auth0-react';
import { Redirect } from 'react-router';
import ThreeDotsWave from './components/screens/Loading/ThreeDotWaves'

let baseUrl = function () {
  let base_url = window.location.origin;
  if (process.env.REACT_APP_API_PORT) {
    let parts = origin.split(":");
    base_url = parts[0] + ":" + parts[1] + ":" + process.env.REACT_APP_API_PORT;
  }
  Constants.setGlobalBaseURL(base_url);
  return base_url;
}();

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      version: "",
      server_metadata: null,
      kelp_errors: {},
      active_error: null, // { botName, level, errorList, index }
    };

    this.setVersion = this.setVersion.bind(this);
    this.fetchServerMetadata = this.fetchServerMetadata.bind(this);
    this.showQuitButton = this.showQuitButton.bind(this);
    this.quit = this.quit.bind(this);
    this.addError = this.addError.bind(this);
    this.addErrorToObject = this.addErrorToObject.bind(this);
    this.removeError = this.removeError.bind(this);
    this.removeErrorsBackend = this.removeErrorsBackend.bind(this);
    this.fetchKelpErrors = this.fetchKelpErrors.bind(this);
    this.findErrors = this.findErrors.bind(this);
    this.setActiveBotError = this.setActiveBotError.bind(this);
    this.hideActiveError = this.hideActiveError.bind(this);
    this._asyncRequests = {};
  }

  // async componentWillMount(){
    // const { user,isAuthenticated, isLoading ,loginWithRedirect, getAccessTokenSilently, handleRedirectCallback } = this.props.auth0;
    // console.log("Printing from componentWillMount app.js 1st", isAuthenticated)
    // if (!isAuthenticated && !window.location.search.includes('code=')) {
      // loginWithRedirect();
    // }
    // if (window.location.search.includes('code=')) {
		// 	// return this.handleRedirectCallback();
    //   return await handleRedirectCallback()
		// }

    // console.log("Printing from componentWillMount app.js 2st", isAuthenticated)


    // if (!isAuthenticated) {
    //     await loginWithRedirect("http://localhost:8000/")
    // }
    // console.log("Printing from componentWillMount app.js 3st", isAuthenticated)

    // console.log("we are not authenticated")

    // if(isAuthenticated){
    //   console.log("we are authenticated")
    //   getAccessTokenSilently().then(token => localStorage.setItem("accessToken", token))
    // }
  // }

  async componentDidMount() {
    // console.log("Printing from componentDidMount app.js", isAuthenticated)


    // // const { getAccessTokenSilently } = this.props.auth0;
    // // getAccessTokenSilently().then(token => localStorage.setItem("accessToken", token))
    // const userIDKey = 'user_id';
    // const Auth0ID = JSON.stringify(user.sub).slice(7).replace('"', '')
    // console.log('the profile ' +JSON.stringify(Auth0ID));
    // localStorage.setItem(userIDKey, Auth0ID);
    // this.setState({user_id: Auth0ID})
    // const accessToken = await getAccessTokenSilently();
    // localStorage.setItem("accessToken", accessToken);
    // this.setState({access_token: accessToken})
    // console.log(isAuthenticated)
    // console.log(this.props.auth0.isAuthenticated)

    
    // if(isAuthenticated){
    //   console.log("we are authenticated in componentDidMount")
    //   getAccessTokenSilently().then(token => localStorage.setItem("accessToken", token))
    // }
    // if(!isAuthenticated){
    //   console.log("we are not authenticated in componentDidMount")
    //   getAccessTokenSilently().then(token => localStorage.setItem("accessToken", token))
    // }

    // if(isAuthenticated){
    //   const userIDKey = 'user_id';
    //   const Auth0ID = JSON.stringify(user.sub).slice(7).replace('"', '')
    //   console.log('the profile ' +JSON.stringify(Auth0ID));
    //   localStorage.setItem(userIDKey, Auth0ID);
    //   const accessToken = await getAccessTokenSilently();
    //   localStorage.setItem("accessToken", accessToken);
    // }
    // while(isLoading && !isAuthenticated){
    //   console.log("priting while authenticating");
    //   if(isAuthenticated){
    //     break;
    //   }
    // }

    
    // this.initializeAuth0()
    // this.getAccessToken();
    this.setVersion()
    this.fetchServerMetadata();

    this.fetchKelpErrors();
    if (!this._fetchKelpErrorsTimer) {
      this._fetchKelpErrorsTimer = setInterval(this.fetchKelpErrors, 500);
    }
  }

  // initializeAuth0 = async () => {
  //   console.log("Printing from initializeAuth0 app.js")
	// 	const auth0Client = await this.props.auth0;
	// 	this.setState({ auth0Client });

	// 	// check to see if they have been redirected after login
	// 	if (window.location.search.includes('code=')) {
	// 		return this.handleRedirectCallback();
	// 	}

	// 	const isAuthenticated = await auth0Client.isAuthenticated();
  //   console.log("Printing from initializeAuth0 app.js", isAuthenticated)
	// 	const user = isAuthenticated ? await auth0Client.getUser() : null;
	// 	this.setState({ isLoading: false, isAuthenticated, user });
	// };

	// handleRedirectCallback = async () => {
  //   const { user,isAuthenticated, isLoading ,loginWithRedirect, getAccessTokenSilently, handleRedirectCallback } = this.props.auth0;

  //   console.log("Printing from handleRedirectCallback app.js")

	// 	this.setState({ isLoading: true });

	// 	// await this.props.auth0.handleRedirectCallback();

  //   console.log("Printing from handleRedirectCallback app.js line 164", user)
    
	// 	// const user = await this.state.auth0Client.getUser();

	// 	this.setState({ user, isAuthenticated: true, isLoading: false });
  //   // console.log("Printing from handleRedirectCallback app.js", isAuthenticated)
  //   console.log("Printing from handleRedirectCallback app.js", user)

	// 	window.history.replaceState({}, document.title, window.location.pathname);
	// };

  // async getAccessToken() {
  //   // const { user, getAccessTokenSilently } = this.props.auth0;
  //   // const token = await getAccessTokenSilently();
  //   // localStorage.setItem("accessToken", token);
  //   // const userIDKey = 'user_id';
  //   // const Auth0ID = JSON.stringify(user.sub).slice(7).replace('"', '')
  //   // console.log('the profile ' +JSON.stringify(Auth0ID));
  //   // localStorage.setItem(userIDKey, Auth0ID);

  //   const { user,isLoading,isAuthenticated, loginWithRedirect, getAccessTokenSilently } = this.props.auth0;
  //   console.log("Printing from render app.js", isAuthenticated)

  //   if(isAuthenticated && localStorage.getItem('accessToken') == null ){
  //     const userIDKey = 'user_id';
  //     // console.log('the profile ' +JSON.stringify(user));
  //     const Auth0ID = JSON.stringify(user.sub).slice(7).replace('"', '')
  //     console.log('the profile ' +JSON.stringify(Auth0ID));
  //     localStorage.setItem(userIDKey, Auth0ID);
  //     // console.log('the profile: ' +JSON.stringify(user.sub).slice(7).replace('"', ''));
  
  //     getAccessTokenSilently().then( token => localStorage.setItem("accessToken", token));
  //   }

  // }

  // handleRedirectCallback = async () => {

	// 	await this.props.auth0.handleRedirectCallback();
	// 	// const user = await this.state.auth0.getUser();
  //   // console.log('the profile ' +JSON.stringify(user));
  //   // const { isAuthenticated } = this.props.auth0;

  //   console.log(this.props.auth0.isAuthenticated)
	// 	// window.history.replaceState({}, document.title, window.location.pathname);
	// };

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

  fetchServerMetadata() {
    var _this = this;
    this._asyncRequests["serverMetadata"] = serverMetadata(baseUrl).then(resp => {
      if (!_this._asyncRequests["serverMetadata"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["serverMetadata"];
      if (!resp["error"]) {
        _this.setState({ server_metadata: resp });
      } else {
        setTimeout(_this.fetchServerMetadata, 30000);
      }
    });
  }

  quit() {
    if (!this.showQuitButton()) {
      console.error("calling quit function when showQuitButton() returned false!");
      return
    }

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

    if (this._fetchKelpErrorsTimer) {
      clearTimeout(this._fetchKelpErrorsTimer);
      this._fetchKelpErrorsTimer = null; 
    }
  }

  addError(backendError) {
    const addResult = this.addErrorToObject(backendError, this.state.kelp_errors);
    const kelp_errors = addResult.kelp_errors;
    const levelErrors = addResult.levelErrors;

    // trigger state change
    if (this.state.active_error === null) {
      this.setState({ "kelp_errors": kelp_errors });
    } else {
      let newState = {
        "kelp_errors": kelp_errors,
        "active_error": this.state.active_error,
      };
      // TODO add support to handle active errors that are not bot type errors
      if (
        backendError.object_type === Constants.ErrorType.bot &&
        backendError.object_name === this.state.active_error.botName &&
        backendError.level === this.state.active_error.level
      ) {
        // update activeErrors when it is affected (either errors or occurrences)
        newState.active_error.errorList = Object.values(levelErrors);
      }
      this.setState(newState);
    }
  }

  addErrorToObject(backendError, kelp_errors) {
    // TODO convert to hashID
    const ID = backendError.message

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
    idError.occurrences.push({
      uuid: backendError.uuid,
      date: backendError.date
    });

    return {
      kelp_errors: kelp_errors,
      levelErrors: levelErrors,
    };
  }

  fetchKelpErrors() {
    var _this = this;
    this._asyncRequests["fetchKelpErrors"] = fetchKelpErrors(baseUrl).then(resp => {
      if (!_this._asyncRequests["fetchKelpErrors"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }
      delete _this._asyncRequests["fetchKelpErrors"];

      if (!resp.kelp_error_list) {
        console.log(resp);
        return
      }

      // parse each error into the kelp_errors object
      let kelp_errors = {};
      resp.kelp_error_list.forEach((ke, index) => {
        // add to kelp_errors
        const addResult = _this.addErrorToObject(ke, kelp_errors);
        kelp_errors = addResult.kelp_errors;
        // we can ignore levelErrors here since we don't need it
      });
      // update state with the new set of errors
      this.setState({ "kelp_errors": kelp_errors });
    });
  }

  removeErrorsBackend(kelpErrorUUIDs) {
    // use first entry as key to async requests
    const networkKey = "removeErrorsBackend_" + kelpErrorUUIDs[0] + "...";
    var _this = this;
    this._asyncRequests[networkKey] = removeKelpErrors(baseUrl, kelpErrorUUIDs).then(resp => {
      if (!_this._asyncRequests[networkKey]) {
        // if it has been deleted it means we don't want to process the result
        return
      }
      delete _this._asyncRequests[networkKey];

      // do nothing here
      // console.log("removed errors: " + JSON.stringify(resp.removed_map));
    });
  }

  findErrors(object_type, object_name, level) {
    const kelp_errors = this.state.kelp_errors;

    if (!kelp_errors.hasOwnProperty(object_type)) {
      return [];
    }
    const botErrors = kelp_errors[object_type];
    
    if (!botErrors.hasOwnProperty(object_name)) {
      return [];
    }
    const namedError = botErrors[object_name];

    if (!namedError.hasOwnProperty(level)) {
      return [];
    }
    const levelErrors = namedError[level];

    // return as an array
    return Object.values(levelErrors);
  }

  removeError(object_type, object_name, level, error) {
    const errorID = error.message;
    let kelp_errors = this.state.kelp_errors;
    let botErrors = kelp_errors[object_type];
    let namedError = botErrors[object_name];
    let levelErrors = namedError[level];

    // save errorUUIDs to be deleted on the backend
    let errorUUIDs = [];
    levelErrors[errorID].occurrences.forEach((o, index) => {
      // add to errorUUIDs
      errorUUIDs.push(o.uuid);
    });
    
    // delete entry for error
    delete levelErrors[errorID];
    // bubble up
    if (Object.keys(levelErrors).length === 0) {
      delete namedError[level];
    }
    if (Object.keys(namedError).length === 0) {
      delete botErrors[object_name];
    }
    if (Object.keys(botErrors).length === 0) {
      delete kelp_errors[object_type];
    }

    let newState = {
      "kelp_errors": kelp_errors,
      "active_error": this.state.active_error,
    };
    // update the error that is now active accordingly
    newState.active_error.errorList = Object.values(levelErrors);
    const wasOnlyError = newState.active_error.errorList.length === 0;
    if (wasOnlyError) {
      newState.active_error = null;
    } else {
      const isLastError = newState.active_error.index > newState.active_error.errorList.length - 1;
      if (isLastError) {
        newState.active_error.index = newState.active_error.errorList.length - 1;
      } 
      // else leave index as-is since we just deleted the index and the new item will now be at the old index (delete in place)
    }

    // send message to backend to remove the error
    this.removeErrorsBackend(errorUUIDs);
    // trigger state change
    this.setState(newState);
  }

  // TODO extend for non-bot type errors later
  setActiveBotError(botName, level, errorList, index) {
    this.setState({
      "active_error": {
        botName: botName,
        level: level,
        errorList: errorList,
        index: index,
      }
    });
  }

  hideActiveError() {
    this.setState({ "active_error": null });
  }

  showQuitButton() {
    // showQuit defaults to false, use this instead of enableKaas so it's more explicit
    return this.state.server_metadata ? !this.state.server_metadata.enable_kaas : false;
  }

  render() {

    const { user,isLoading,isAuthenticated, loginWithRedirect, getAccessTokenSilently } = this.props.auth0;


    // if(!isAuthenticated) { // if your component doesn't have to wait for an async action, remove this block 
    // return null; // render null when app is not ready
    // }
  
    // const { user } = this.props.auth0;
    // console.log("printing from render",user.name);
    // return <div>Hello {user.name}</div>;
    // if (!isAuthenticated && !window.location.search.includes('code=')) {
    //   loginWithRedirect();
    // }
  
    // if (isLoading) {
    //   return ;
    // }

    // const { user,isLoading,isAuthenticated, loginWithRedirect, getAccessTokenSilently } = this.props.auth0;

    // construction of metricsTracker in server_amd64.go (isTestnet) needs to logically match this variable
    // we use the state because that is updated from the /serverMetadata endpoint
    const enablePubnetBots = this.state.server_metadata ? !this.state.server_metadata.disable_pubnet : false;

    let quitButton = "";
    if (this.showQuitButton()) {
      quitButton = (
        <Button
          eventName="main-quit"
          className={styles.quit}
          size="small"
          onClick={this.quit}
        >
          Quit
        </Button>
      );
    }

    let banner = (<div className={styles.banner}>
      {quitButton}
      Kelp GUI (beta) v1.0.0-rc2
    </div>);

    const removeBotError = this.removeError.bind(this, Constants.ErrorType.bot);
    const findBotErrors = this.findErrors.bind(this, Constants.ErrorType.bot);

    return (
      <div>
        <div>{banner}</div>
        <Router>
        <LoginRedirect/>
        {/* {!isAuthenticated && <Redirect exact from="/" to="/loading" />} */}
          {/* <Route  exact path="/loading" render={<ThreeDotsWave/>}/> */}
          <Header version={this.state.version}/>
          <Route exact path="/"
            render={(props) => <Bots {...props} baseUrl={baseUrl} enablePubnetBots={enablePubnetBots} activeError={this.state.active_error} setActiveError={this.setActiveBotError} hideActiveError={this.hideActiveError} addError={this.addError} removeError={removeBotError} findErrors={findBotErrors}/>}
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
        <Welcome quitFn={this.quit} showQuitButton={this.showQuitButton()}/>
      </div>
    );
  }
}

export default withAuth0(App);
