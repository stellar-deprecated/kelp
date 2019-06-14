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
      botName={botName}
      configData={this.state.configData}
      saveFn={this.saveEdit}
      saveText="Save Bot Updates"
      />);
  }
}

export default NewBot;
