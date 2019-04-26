import React, { Component } from 'react';
import './FieldGroup.css';

class FieldGroup extends Component {

  render() {
    return (
      <div>
        {this.props.children}
      </div>
    );
  }
}

export default FieldGroup;