import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Button from '../Button/Button';


class StartStop extends Component {
  static defaultProps = {
    isRunning: false,
    onClick: () => {},
  }

  static propTypes = {
    isRunning: PropTypes.bool,
    onClick: PropTypes.func,
  };

  render() {
    if(this.props.isRunning){
      return (
        <Button icon="stop" size="small" variant="stop" onClick={this.props.onClick}>Stop</Button>
      )  
    }
    else {
      return (
        <Button icon='start' size="small" variant="start" onClick={this.props.onClick}>Start</Button>
      )
    }
  }
}

export default StartStop;