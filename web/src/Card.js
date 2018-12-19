import React, {Component} from 'react';
import RefreshIcon from '@material-ui/icons/Refresh';
import IconButton from '@material-ui/core/IconButton';
import {withTheme} from '@material-ui/core/styles';
import tinycolor from 'tinycolor2'

const css = (theme) => {
  const background = theme.palette.background.default;

  const borderColor = tinycolor(theme.palette.text.primary).setAlpha(.5);

  const styles = {
    wrapper: {
      color: theme.palette.text.primary,
      background: background,
      display: "flex",
      flexDirection: "column",
      borderRadius: "6px",
      margin: "20px",
      border: "solid 1px " + borderColor
    },
    topBar: {
      display: "flex",
      alignItems: "center",
      fontSize: "1.4em",
      textTransform: "uppercase",
      borderRadius: "6px",
      padding: "10px 16px"
    },
    titleText: {
      flex: "1 0 auto"
    }
  }

  return styles
}

class Card extends Component {
  render() {
    const {theme} = this.props;

    const styles = css(theme)

    return (<div style={styles.wrapper}>
      <div style={styles.topBar}>
        <div style={styles.titleText}>
          {this.props.title}
        </div>
        <IconButton aria-label="Refresh" onClick={this.props.refresh}>
          <RefreshIcon/>
        </IconButton>
      </div>
      <div style={styles.offerTables}>
        {this.props.children}
      </div>
    </div>)
  }
}

export default withTheme()(Card);
