import React from "react";
import {withTheme} from '@material-ui/core/styles';

class TaskOffers extends React.Component {
  render() {
    // const {theme} = this.props;
    // const primaryText = theme.palette.text.primary;

    return (<div style={this.props.tabStyles.tabContainer}>
      <div>log</div>
    </div>);
  }
}

export default withTheme()(TaskOffers);
