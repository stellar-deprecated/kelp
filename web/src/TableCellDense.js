import TableCell from '@material-ui/core/TableCell';
import {withTheme, withStyles} from '@material-ui/core/styles';

const TableCellDense = withStyles(theme => ({
  head: {
    backgroundColor: "rgba(0,0,0,.3)",
    color: theme.palette.common.white
  },
  body: {
    fontSize: 14
  }
}))(TableCell)

TableCellDense.defaultProps = {
  // ["default","checkbox","dense","none"].
  // dense wasn't enough, so setting padding above
  padding: 'dense'
}

const TableCellMicro = withStyles(theme => ({
  head: {
    backgroundColor: "rgba(0,0,0,.3)",
    color: theme.palette.common.white,
    padding: "0 8px"
  },
  body: {
    fontSize: 14,
    padding: "0 8px"
  }
}))(TableCell)

const Dense = withTheme()(TableCellDense)
const Micro = withTheme()(TableCellMicro)
export {
  Dense,
  Micro
}
