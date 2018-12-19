import React, {Component} from 'react';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import {withTheme} from '@material-ui/core/styles';
import ClearIcon from '@material-ui/icons/Clear';
import IconButton from '@material-ui/core/IconButton';
import TaskDialog from './taskDialog/TaskDialog.js';
import {Dense} from './TableCellDense.js'

const styles = {
  table: {
    maxWidth: "900px"
  }
}

class TasksTable extends Component {
  state = {
    selected: "",
    dialogOpen: false,
    dialogData: null
  }

  render() {
    let offerTable = ''
    const offers = this.props.tasks
    // const {theme} = this.props;

    const TableContents = (props) => {
      if (offers.length > 0) {
        return offers.map((row) => {
          return (<TableRow hover={true} onClick={event => this.handleRowClick(event, row)} key={row.pid}>
            <Dense>{row.pid}</Dense>
            <Dense>{row.name}</Dense>
            <Dense>{row.cmd}</Dense>
            <Dense>
              <IconButton aria-label="Kill" onClick={(e) => this.killClick(e, row)}>
                <ClearIcon/>
              </IconButton>
            </Dense>
          </TableRow>)
        })
      }

      return (<TableRow hover={true} onClick={event => this.handleClick(event, 'refresh')}>
        <Dense>No bots. Click to refresh.</Dense>
        <Dense></Dense>
        <Dense></Dense>
        <Dense></Dense>
      </TableRow>)
    }

    offerTable = (<Table style={styles.table}>
      <TableHead>
        <TableRow >
          <Dense>PID</Dense>
          <Dense>Name</Dense>
          <Dense>Command</Dense>
          <Dense>Kill</Dense>
        </TableRow>
      </TableHead>
      <TableBody>
        <TableContents/>
      </TableBody>
    </Table>)

    return (<div >
      <TaskDialog close={this.dialogClose} data={this.state.dialogData} open={this.state.dialogOpen}></TaskDialog>
      {offerTable}
    </div>)
  }

  dialogClose = (okClicked) => {
    this.setState({dialogOpen: false})
  }

  killClick = (event, row) => {
    // clicking on the table row will also fire unless we stop it here
    event.preventDefault()
    event.stopPropagation()

    this.props.handler(row)
  }

  handleRowClick = (event, row) => {
    const {selected} = this.state;
    let newPid = row.pid

    // turn off if clicked twice
    if (row.pid === selected) {
      newPid = ''
    }

    this.setState({
      selected: newPid,
      dialogOpen: true,
      dialogData: {
        project: row.project
      }
    })
  }

  handleClick = (event, id) => {
    if (id === 'refresh') {
      this.props.refresh()
    }
  }
}

export default withTheme()(TasksTable);
