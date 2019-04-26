import React, { Component } from 'react';
import './ScreenTitle.css';

class ScreenTitle extends Component {


  render() {
    return (
      <h1>{this.props.children}</h1>
    );
  }
}

export default ScreenTitle;