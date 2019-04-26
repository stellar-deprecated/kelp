import React, { Component } from 'react';

import './BackButton.css';

import arrowBackIcon from '../../../assets/images/ico-arrow-back.svg';


class BackButton extends Component {

  render() {
    return (
      <button>
        <img src={arrowBackIcon}/>
      </button>
    );
  }
}

export default BackButton;