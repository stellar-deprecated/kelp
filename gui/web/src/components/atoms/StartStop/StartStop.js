import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Button from '../Button/Button';
import Constants from '../../../Constants';

class StartStop extends Component {
  static defaultProps = {
    state: Constants.BotState.initializing,
    onClick: () => {},
  }

  static propTypes = {
    state: PropTypes.string,
    onClick: PropTypes.func,
  };

  render() {
    let icon = "";
    let variant = "";
    let text = ""
    if (this.props.state === Constants.BotState.running) {
      icon = "stop";
      variant = "stop";
      text = "Stop";
    } else {
      icon = "start";
      variant = "start";
      text = "Start";
    }
    let disabled = this.props.state === Constants.BotState.initializing || this.props.state === Constants.BotState.stopping;

    return (<Button icon={icon} size="small" variant={variant} onClick={this.props.onClick} disabled={disabled}>{text}</Button>);
  }
}

export default StartStop;