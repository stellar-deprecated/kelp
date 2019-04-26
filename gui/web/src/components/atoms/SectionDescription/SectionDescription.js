import React, { Component } from 'react';
import './SectionDescription.css';

class SectionDescription extends Component {

  render() {
    return (
      <p>{this.props.children}</p>
    );
  }
}

export default SectionDescription;