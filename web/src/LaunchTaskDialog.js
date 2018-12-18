import React, {Component} from 'react';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import InputLabel from '@material-ui/core/InputLabel';
import FilledInput from '@material-ui/core/FilledInput';
import {withStyles} from '@material-ui/core/styles';

const styles = theme => ({
  root: {
    display: 'flex',
    flexDirection: 'column',
    flexWrap: 'wrap'
  },
  formControl: {
    margin: theme.spacing.unit,
    minWidth: 120
  },
  selectEmpty: {
    marginTop: theme.spacing.unit * 2
  }
});

class LaunchTaskDialog extends Component {
  state = {
    project: 'default'
  }

  handleChange = name => event => {
    this.setState({[name]: event.target.value});
  };

  render() {
    const {classes} = this.props;

    return (<div >
      <Dialog maxWidth={false} open={this.props.open} onClose={(e) => this.handleClick('dialogCancel')}>

        <DialogContent>

          <div>
            <div className={classes.root}>
              <div>New {this.props.taskId}</div>
              <FormControl variant="filled" className={classes.formControl}>
                <InputLabel htmlFor="filled-project-native-simple">Token</InputLabel>
                <Select native={true} value={this.state.project} onChange={this.handleChange('project')} input={<FilledInput name = "project" id = "filled-project-native-simple" />}>
                  <option value={'default'}>default</option>
                  <option value={'kelpr'}>kelpr</option>
                </Select>
              </FormControl>
            </div>
          </div>
        </DialogContent>
        <DialogActions>
          <Button variant="contained" color="secondary" onClick={(e) => this.handleClick('dialogCancel')}>
            Cancel
          </Button>
          <Button variant="contained" color="primary" onClick={(e) => this.handleClick('dialogOK')}>
            Start
          </Button>
        </DialogActions>
      </Dialog>
    </div>);
  }

  // binds this for callback without .bind(this)
  handleClick = (id) => {
    switch (id) {
      case 'dialogOK':
        this.props.close(true, this.state.project)
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
export default withStyles(styles)(LaunchTaskDialog);
