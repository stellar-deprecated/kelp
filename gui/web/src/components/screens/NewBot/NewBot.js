import React, { Component } from 'react';
import Form from '../../molecules/Form/Form';

class NewBot extends Component {
  render() {
    return (
      <Form router={this.props.history}/>
    );
  }
}

export default NewBot;
