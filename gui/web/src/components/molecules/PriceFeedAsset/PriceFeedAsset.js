import React, { Component } from 'react';
import PropTypes from 'prop-types';
import PriceFeedTitle from '../PriceFeedTitle/PriceFeedTitle';
import PriceFeedSelector from '../PriceFeedSelector/PriceFeedSelector';

class PriceFeedAsset extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isLoading: true,
      price: null
    };
    this.fetchPrice = this.fetchPrice.bind(this);

    this._asyncRequests = {};
  }

  static propTypes = {
    title: PropTypes.string,
    type: PropTypes.string,
    feed_url: PropTypes.string,
  };

  componentDidMount() {
    this.fetchPrice();
  }

  fetchPrice() {
    // if (this._asyncRequests["price"]) {
    //   return
    // }
    // 
    // var _this = this;
    // this._asyncRequests["price"] = getPrice(this.props.baseUrl, this.props.type, this.props.feed_url).then(resp => {
    //   _this._asyncRequests["price"] = null;
    //   _this.setState({
    //     price: resp.price,
    //   });
    // });
    this.setState({
      isLoading: true,
    });

    // temp
    setTimeout(() => {
      this.setState({
        isLoading: false,
        price: 0.115,
      });
    }, 5000);
  }

  // componentWillUnmount() {
  //   if (this._asyncRequests["price"]) {
  //     this._asyncRequests["price"].cancel();
  //     this._asyncRequests["price"] = null;
  //   }
  // }

  render() {
    if (!this.state.price || this.state.isLoading) {
      return (
        <div>
          <PriceFeedTitle
            label={this.props.title}
            loading={true}
            fetchPrice={this.fetchPrice}
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
          fetchPrice={this.fetchPrice}
          />
        <PriceFeedSelector/>
      </div>
    );
  }
}

export default PriceFeedAsset;