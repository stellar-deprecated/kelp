import React, { Component } from 'react';
import PropTypes from 'prop-types';
import PriceFeedTitle from '../PriceFeedTitle/PriceFeedTitle';
import PriceFeedSelector from '../PriceFeedSelector/PriceFeedSelector';
import fetchPrice from '../../../kelp-ops-api/fetchPrice';

class PriceFeedAsset extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isLoading: true,
      price: null
    };
    this.queryPrice = this.queryPrice.bind(this);

    this._asyncRequests = {};
  }

  static propTypes = {
    baseUrl: PropTypes.string,
    title: PropTypes.string,
    type: PropTypes.string,
    feed_url: PropTypes.string,
  };

  componentDidMount() {
    this.queryPrice();
  }

  queryPrice() {
    if (this._asyncRequests["price"]) {
      return
    }
    
    var _this = this;
    this._asyncRequests["price"] = fetchPrice(this.props.baseUrl, this.props.type, this.props.feed_url).then(resp => {
      _this._asyncRequests["price"] = null;
      _this.setState({
        isLoading: false,
        price: resp.price,
      });
    });
    this.setState({
      isLoading: true,
    });
  }

  componentWillUnmount() {
    if (this._asyncRequests["price"]) {
      this._asyncRequests["price"].cancel();
      this._asyncRequests["price"] = null;
    }
  }

  render() {
    if (!this.state.price || this.state.isLoading) {
      return (
        <div>
          <PriceFeedTitle
            label={this.props.title}
            loading={true}
            fetchPrice={this.queryPrice}
            />
          <PriceFeedSelector/>
        </div>
      );  
    }

    return (
      <div>
        <PriceFeedTitle
          label={this.props.title}
          loading={false}
          price={this.state.price}
          fetchPrice={this.queryPrice}
          />
        <PriceFeedSelector/>
      </div>
    );
  }
}

export default PriceFeedAsset;