import React from "react";
import ToggleButton from "@material-ui/lab/ToggleButton";
import ToggleButtonGroup from "@material-ui/lab/ToggleButtonGroup";
import {withTheme, withStyles} from '@material-ui/core/styles';

// customizing button colors
const TButton = withStyles({
  root: {},
  selected: {
    background: "rgb(10, 130, 250)",
    color: 'white'
  }
})(ToggleButton);

class NetworkSwitch extends React.Component {
  state = {
    buttonValue: 'test'
  };

  handleClick = (event, buttonValue) => {
    if (buttonValue && buttonValue.length > 0) {
      this.setState({buttonValue})
      this.props.changed(buttonValue)
    }
  }

  render() {
    const {theme} = this.props;
    const primaryText = theme.palette.text.primary;

    const styles = {
      toggleContainer: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        margin: "10px"
      },
      url: {
        marginTop: "8px",
        color: primaryText
      }
    }

    const {buttonValue} = this.state;
    let url = 'https://horizon.stellar.org'

    if (buttonValue === 'test') {
      url = 'https://horizon-testnet.stellar.org'
    }

    return (<div style={styles.toggleContainer}>
      <ToggleButtonGroup value={buttonValue} exclusive={true} onChange={this.handleClick}>
        <TButton value="test">Test</TButton>
        <TButton value="public">Public</TButton>
      </ToggleButtonGroup>
      <div style={styles.url}>{url}</div>
    </div>);
  }
}

export default withTheme()(NetworkSwitch);
