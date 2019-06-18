import React, { Component } from 'react';
import Form from '../../molecules/Form/Form';
import genBotName from '../../../kelp-ops-api/genBotName';
import getBotConfig from '../../../kelp-ops-api/getBotConfig';

class NewBot extends Component {
  constructor(props) {
    super(props);
    this.state = {
      newBotName: null,
      configData: null,
    };

    this.loadBotName = this.loadBotName.bind(this);
    this.saveNew = this.saveNew.bind(this);
    this.saveEdit = this.saveEdit.bind(this);
    this.loadSampleConfigData = this.loadSampleConfigData.bind(this);
    this.loadBotConfigData = this.loadBotConfigData.bind(this);
    this.onChangeForm = this.onChangeForm.bind(this);
    this.updateUsingDotNotation = this.updateUsingDotNotation.bind(this);

    this._asyncRequests = {};
  }

  loadBotName() {
    if (this.state.newBotName) {
      return
    }

    var _this = this;
    this._asyncRequests["botName"] = genBotName(this.props.baseUrl).then(resp => {
      _this._asyncRequests["botName"] = null;
      _this.setState({
        newBotName: resp,
      });
    });
  }

  componentWillUnmount() {
    if (this._asyncRequests["botName"]) {
      this._asyncRequests["botName"].cancel();
      this._asyncRequests["botName"] = null;
    }

    if (this._asyncRequests["botConfig"]) {
      this._asyncRequests["botConfig"].cancel();
      this._asyncRequests["botConfig"] = null;
    }
  }

  saveNew(configData) {
    return null;
  }

  saveEdit(configData) {
    return null;
  }

  loadSampleConfigData() {

  }

  loadBotConfigData(botName) {
    var _this = this;
    this._asyncRequests["botConfig"] = getBotConfig(this.props.baseUrl, botName).then(resp => {
      _this._asyncRequests["botConfig"] = null;
      _this.setState({
        configData: resp,
      });
    });
  }

  updateUsingDotNotation(obj, path, newValue) {
    // update correct value by converting from dot notation string
    let parts = path.split('.');

    // maintain reference to original object by creating copy
    let current = obj;

    // fetch the object that contains the field we want to update
    for (let i = 0; i < parts.length - 1; i++) {
      current = current[parts[i]];
    }

    // update the field
    current[parts[parts.length-1]] = newValue;
  }

  onChangeForm(statePath, event, mergeUpdateInstructions) {
    // make copy of current state
    let updateJSON = Object.assign({}, this.state);

    this.updateUsingDotNotation(updateJSON.configData, statePath, event.target.value);

    // merge in any additional updates
    if (mergeUpdateInstructions) {
      let keys = Object.keys(mergeUpdateInstructions)
      for (let i = 0; i < keys.length; i++) {
        let dotNotationKey = keys[i];
        let fn = mergeUpdateInstructions[dotNotationKey];
        let newValue = fn(event.target.value);
        if (newValue != null) {
          this.updateUsingDotNotation(updateJSON.configData, dotNotationKey, newValue);
        }
      }
    }

    // set state for the full state object
    this.setState(updateJSON);
  }

  render() {
    if (this.props.location.pathname === "/new") {
      this.loadBotName();
      if (!this.state.configData) {
        this.loadSampleConfigData();
        return (<div>Fetching sample config file</div>);
      }
      return (<Form
        router={this.props.history}
        title="New Bot"
        onChange={this.onChangeForm}
        botName={this.state.newBotName}
        configData={this.state.configData}
        saveFn={this.saveNew}
        saveText="Create Bot"
        />);
    } else if (this.props.location.pathname !== "/edit") {
      console.log("invalid path: " + this.props.location.pathname);
      return "";
    }

    if (this.props.location.search.length === 0) {
      console.log("no search params provided to '/edit' route");
      return "";
    }

    let searchParams = new URLSearchParams(this.props.location.search.substring(1));
    let botNameEncoded = searchParams.get("bot_name");
    if (!botNameEncoded) {
      console.log("no botName param provided to '/edit' route");
      return "";
    }

    let botName = decodeURIComponent(botNameEncoded);
    if (!this.state.configData) {
      this.loadBotConfigData(botName);
      return (<div>Fetching config file for bot: {botName}</div>);
    }
    return (<Form 
      router={this.props.history}
      title="Edit Bot"
      onChange={this.onChangeForm}
      botName={botName}
      configData={this.state.configData}
      saveFn={this.saveEdit}
      saveText="Save Bot Updates"
      />);
  }
}

export default NewBot;
