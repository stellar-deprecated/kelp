import React, {Component} from 'react';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import {withTheme} from '@material-ui/core/styles';
import {Micro} from './TableCellDense.js'

const styles = {
  table: {
    maxWidth: "900px"
  },

  offerTables: {
    display: 'flex'
  },
  spacer: {
    width: '10px'
  }
}

class OffersTable extends Component {
  tableJSX(styles, offers, tableStyles) {
    return (<Table style={Object.assign({}, tableStyles, styles.table)}>
      <TableHead>
        <TableRow>
          <Micro>Buying</Micro>
          <Micro>Selling</Micro>
          <Micro>Amount</Micro>
          <Micro>Price</Micro>
        </TableRow>
      </TableHead>
      <TableBody>
        {
          offers.map((row, i) => {
            return (<TableRow key={i}>
              <Micro >{row.buying}</Micro>
              <Micro >{row.selling}</Micro>
              <Micro >{row.amount}</Micro>
              <Micro >{row.price}</Micro>
            </TableRow>);
          })
        }
      </TableBody>
    </Table>)
  }

  render() {
    const buyOffers = this.props.offers.buyOffers
    const sellOffers = this.props.offers.sellOffers
    // const {theme} = this.props;

    let sellTable = ''
    if (sellOffers && sellOffers.length > 0) {
      sellTable = this.tableJSX(styles, sellOffers, {background: 'rgb(255,100,100)'})
    }

    let buyTable = ''
    if (buyOffers && buyOffers.length > 0) {
      buyTable = this.tableJSX(styles, buyOffers, {background: "rgb(100,255,100)"})
    }

    return (<div style={styles.offerTables}>
      {sellTable}
      <div style={styles.spacer}/> {buyTable}
    </div>)
  }
}

export default withTheme()(OffersTable);
