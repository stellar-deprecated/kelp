import React, { Component } from 'react';
import Form from '../../molecules/Form/Form';
import genBotName from '../../../kelp-ops-api/genBotName';

class NewBot extends Component {
  constructor(props) {
    super(props);
    this.state = {
      newBotName: null,
      configData: {},
    };

    this.loadBotName = this.loadBotName.bind(this);
    this.saveNew = this.saveNew.bind(this);
    this.saveEdit = this.saveEdit.bind(this);

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
  }

  saveNew() {
    return null;
  }

  saveEdit() {
    return null;
  }

  render() {
    if (this.props.location.pathname === "/new") {
      this.loadBotName();
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
