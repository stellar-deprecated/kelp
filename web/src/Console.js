import React, {Component} from 'react';
import {withTheme} from '@material-ui/core/styles';
import ButtonBar from './ButtonBar.js'
import DialogButton from './DialogButton.js'
import KelpTalk from './KelpTalk.js'
import TextField from '@material-ui/core/TextField';
import tinycolor from 'tinycolor2'

class Console extends Component {
  state = {
    kelpParams: 'help trade',
    dialogOpen: false,
    contents: ''
  }

  constructor(props) {
    super(props)

    KelpTalk.on('console', (value) => {
      this.setState({"contents": value})
    })
  }

  render() {
    const {theme} = this.props;
    // const te = theme.palette.background.default;
    const borderColor = tinycolor(theme.palette.text.primary).setAlpha(.25);

    const styles = {
      textButtonWrapper: {
        display: "flex",
        justifyContent: "center",
        flexWrap: "wrap",
        margin: "10px 0"
      },
      responseWrapper: {
        overflow: 'auto',
        flex: "1 1 auto",
        height: "400px",
        background: "rgb(55, 65, 55)",
        border: "solid 1px " + borderColor
      },
      responseText: {
        fontFamily: "monospace",
        whiteSpace: "pre",
        padding: "10px",
        // had to set width to get it to shrink with wide text
        width: "1px",
        color: 'rgb(100,255,100)'
      }
    }

    const buttons = [
      {
        title: 'Kelp',
        id: 'kelp'
      }, {
        title: 'Help',
        id: 'help'
      }, {
        title: 'Version',
        id: 'version'
      }, {
        title: 'Exchanges',
        id: 'exchanges'
      }, {
        title: 'Strategies',
        id: 'strategies'
      }, {
        title: 'Config',
        id: 'config'
      }, {
        title: 'Custom...',
        id: 'custom'
      }
    ]

    return (<div>
      <div style={styles.responseWrapper}>
        <div style={styles.responseText}>
          {this.state.contents}
        </div>
      </div>

      <ButtonBar buttons={buttons} click={this.handleClick}/>
      <DialogButton open={this.state.dialogOpen} click={this.handleLaunchDialog} title='Launch Kelp' okButton='OK' message='Using parameters'>
        <div>
          <TextField style={styles.textField} id="params" label="Parameters to kelp" value={this.state.kelpParams} onChange={this.handleChange('kelpParams')} margin="normal"/>
        </div>
      </DialogButton>

    </div>);
  }

  handleLaunchDialog = (id) => {
    switch (id) {
      case 'ok':
        this.setState({dialogOpen: false})
        KelpTalk.put('params', {Kelp: this.state.kelpParams})
        break
      case 'cancel':
        this.setState({dialogOpen: false})
        break
      default:
        break
    }
  }

  handleChange = name => event => {
    this.setState({[name]: event.target.value});
  };

  handleClick = (id) => {
    if (id === 'custom') {
      this.setState({dialogOpen: true})
    } else {
      switch (id) {
        case 'kelp':
          KelpTalk.get('')
          break
        default:
          KelpTalk.get(id)
          break
      }
    }
  }
}

export default withTheme()(Console);
