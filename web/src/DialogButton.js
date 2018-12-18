import React, {Component} from 'react';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogActions from '@material-ui/core/DialogActions';

const styles = {
  wrapper: {
    margin: "8px"
  }
}

export default class DialogButton extends Component {
  render() {
    return (<div style={styles.wrapper}>
      <Dialog fullWidth={true} open={this.props.open} onClose={(e) => this.handleClick('dialogCancel')}>
        <DialogTitle>
          {this.props.title}
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            {this.props.message}
          </DialogContentText>
          {this.props.children}
        </DialogContent>
        <DialogActions>
          <Button variant="contained" onClick={(e) => this.handleClick('dialogCancel')}>
            Cancel
          </Button>
          <Button variant="contained" color="primary" onClick={(e) => this.handleClick('dialogOK')}>
            {this.props.okButton}
          </Button>
        </DialogActions>
      </Dialog>
    </div>);
  }

  // binds this for callback without .bind(this)
  handleClick = (id) => {
    let action = ''

    switch (id) {
      case 'dialogOK':
        action = "ok"
        break
      case 'dialogCancel':
        action = "cancel"
        break
      default:
        console.log(id, 'not handled')
        break
    }

    if (action.length > 0) {
      if (this.props.click) {
        this.props.click(action)
      }
    }
  }
}
