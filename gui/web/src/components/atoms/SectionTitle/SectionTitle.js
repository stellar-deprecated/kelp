import React, { Component } from 'react';
import './SectionTitle.css';

class SectionTitle extends Component {

  render() {
    return (
      <h3>{this.props.children}</h3>
    );
  }
}

export default SectionTitle;