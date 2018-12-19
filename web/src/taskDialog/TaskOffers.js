import React from "react";
import {withTheme} from '@material-ui/core/styles';
import OffersTable from '../OffersTable.js'
import KelpTalk from '../KelpTalk.js'

class TaskOffers extends React.Component {
  state = {
    kelpOffers: []
  }

  constructor(props) {
    super(props)

    KelpTalk.on('updated', (key, value) => {
      switch (key) {
        case 'offers':
          const json = this.offersToJSON(value)

          this.setState({kelpOffers: json})
          break
        default:
          break
      }
    })
  }

  componentDidMount() {
    KelpTalk.get('offers', {project: this.props.project})
  }

  render() {
    // const {theme} = this.props;
    // const primaryText = theme.palette.text.primary;

    return (<div style={this.props.tabStyles.tabContainer}>
      <OffersTable offers={this.state.kelpOffers}></OffersTable>
    </div>);
  }

  convertOffers(offers) {
    const result = []

    for (const x of offers) {
      const data = {}

      data['price'] = x.price
      data['amount'] = x.amount

      if (x.buying) {
        if (x.buying.asset_type === 'native') {
          data['buying'] = 'XLM'
        } else {
          data['buying'] = x.buying.asset_code
        }
      }

      if (x.selling) {
        if (x.selling.asset_type === 'native') {
          data['selling'] = 'XLM'
        } else {
          data['selling'] = x.selling.asset_code
        }
      }

      result.push(data)
    }

    return result
  }

  offersToJSON(offers) {
    const convertedOffers = this.convertOffers(offers)
    const buyOffers = []
    const sellOffers = []

    for (const offer of convertedOffers) {
      if (offer.buying === 'XLM') {
        sellOffers.push(offer)
      } else {
        buyOffers.push(offer)
      }
    }

    return {buyOffers, sellOffers}
  }
}

export default withTheme()(TaskOffers);
