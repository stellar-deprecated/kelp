import React, {Component} from 'react';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';
import AppBar from '@material-ui/core/AppBar';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import PropTypes from 'prop-types';
import Typography from '@material-ui/core/Typography';
import TaskOffers from './TaskOffers.js';
import TaskLog from './TaskLog.js';
import TaskConfig from './TaskConfig.js';
import TaskManage from './TaskManage.js';

function TabContainer(props) {
  return (<Typography component="div" style={{
      padding: 8 * 3
    }}>
    {props.children}
  </Typography>);
}

TabContainer.propTypes = {
  children: PropTypes.node.isRequired
};

class TaskDialog extends Component {
  state = {
    value: 0
  }

  handleChange = (event, value) => {
    this.setState({value});
  };

  render() {
    const {value} = this.state;
    // const {theme} = this.props;
    // const primaryText = theme.palette.text.primary;

    const tabStyles = {
      tabContainer: {
        display: "flex",
        minHeight: "400px",
        margin: "10px"
      }
    }

    return (<div >
      <Dialog maxWidth={false} open={this.props.open} onClose={(e) => this.handleClick('dialogCancel')}>
        <DialogContent>
          <div>
            <AppBar position="static" color="default">
              <Tabs fullWidth={true} value={value} onChange={this.handleChange}>
                <Tab label="Manage"/>
                <Tab label="Offers"/>
                <Tab label="Config"/>
                <Tab label="Log"/>
              </Tabs>
            </AppBar>
            {value === 0 && <TaskManage project={this.props.data && this.props.data.project} tabStyles={tabStyles}/>}
            {value === 1 && <TaskOffers project={this.props.data && this.props.data.project} tabStyles={tabStyles}/>}
            {value === 2 && <TaskConfig tabStyles={tabStyles}/>}
            {value === 3 && <TaskLog tabStyles={tabStyles}/>}
          </div>
        </DialogContent>
        <DialogActions>
          <Button variant="contained" color="primary" onClick={(e) => this.handleClick('dialogOK')}>
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </div>);
  }

  // binds this for callback without .bind(this)
  handleClick = (id) => {
    switch (id) {
      case 'dialogOK':
        this.props.close(true)
        break
      case 'dialogCancel':
        this.props.close(false)
        break
      default:
        console.log(id, 'not handled')
        break
    }
  }
}

export default TaskDialog
