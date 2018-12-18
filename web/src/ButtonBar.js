import React, {Component} from 'react';
import Button from '@material-ui/core/Button';
import {withTheme, withStyles} from '@material-ui/core/styles';

// customizing button colors
const TButton = withStyles({
  root: {
    margin: "0 8px",
    color: 'steelblue'
  }
})(Button);

class ButtonBar extends Component {
  render() {
    // const {theme} = this.props

    const styles = {
      wrapper: {
        display: "flex",
        margin: '10px 0',
        justifyContent: "center",
        flexWrap: 'wrap'
      }
    }
    const buttons = this.props.buttons

    return (<div style={styles.wrapper}>
      {
        buttons.map((row) => {
          return (<TButton variant='text' aria-label={row.title} onClick={e => this.props.click(row.id)} key={row.id}>
            {row.title}
          </TButton>)
        })
      }

    </div>)
  }
}

export default withTheme()(ButtonBar);
