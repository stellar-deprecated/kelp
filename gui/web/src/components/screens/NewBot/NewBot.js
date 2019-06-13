import React, { Component } from 'react';
import Form from '../../molecules/Form/Form';

class NewBot extends Component {
  render() {
    if (this.props.location.pathname === "/new") {
      return (<Form router={this.props.history}/>);
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
    return (
      <Form router={this.props.history}/>
    );
  }
}

export default NewBot;
