import React from "react";
import Button from '@material-ui/core/Button';
import KelpTalk from '../KelpTalk.js'

import {withTheme, withStyles} from '@material-ui/core/styles';

// customizing button colors
const TButton = withStyles({
  root: {
    margin: "8px 0"
  }
})(Button);

const styles = {
  container: {
    flexDirection: "row"
  },
  buttonColumn: {
    display: "flex",
    flexDirection: 'column',
    flex: "0 0 auto"
  },
  infoColumn: {
    padding: "20px",
    display: "flex",
    flexDirection: 'column',
    flex: "1 0 auto"
  }
}

class TaskOffers extends React.Component {
  render() {
    // const {theme} = this.props;
    // const primaryText = theme.palette.text.primary;

    return (<div style={Object.assign({}, styles.container, this.props.tabStyles.tabContainer)}>
      <div style={styles.buttonColumn}>
        <TButton variant="contained" onClick={(e) => this.handleClick('stop')}>
          Stop
        </TButton>
        <TButton variant="contained" onClick={(e) => this.handleClick('start')}>
          Start
        </TButton>
        <TButton variant="contained" onClick={(e) => this.handleClick('delete')}>
          Delete Offers
        </TButton>
        <TButton variant="contained" onClick={(e) => this.handleClick('trustlines')}>
          Add Trustlines
        </TButton>
      </div>
      <div style={styles.infoColumn}>
        <div>stuff here</div>
      </div>
    </div>);
  }

  handleClick = (id) => {
    switch (id) {
      case 'stop':
        break
      case 'start':
        break
      case 'delete':
        KelpTalk.get('delete', {project: this.props.project})
        break
      case 'trustlines':
        break
      default:
        console.log('switch not handled:', id)
        break
    }
  }
}

export default withTheme()(TaskOffers);
