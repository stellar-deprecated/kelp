import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './PriceFeedAsset.module.scss';
import PriceFeedTitle from '../PriceFeedTitle/PriceFeedTitle';
import PriceFeedSelector from '../PriceFeedSelector/PriceFeedSelector';
import fetchPrice from '../../../kelp-ops-api/fetchPrice';
import LoadingAnimation from '../../atoms/LoadingAnimation/LoadingAnimation';

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
    onChange: PropTypes.func,
    onLoadingPrice: PropTypes.func,
    onNewPrice: PropTypes.func,
    optionsMetadata: PropTypes.object,
  };

  componentDidMount() {
    this.queryPrice();
  }

  componentDidUpdate(prevProps) {
    if (
      prevProps.baseUrl !== this.props.baseUrl ||
      prevProps.type !== this.props.type ||
      prevProps.feed_url !== this.props.feed_url
    ) {
      this.queryPrice();
    }
  }

  queryPrice() {
    if (this._asyncRequests["price"]) {
      return
    }
    
    this.setState({
      isLoading: true,
    });
    this.props.onLoadingPrice();

    var _this = this;
    this._asyncRequests["price"] = fetchPrice(this.props.baseUrl, this.props.type, this.props.feed_url).then(resp => {
      if (!_this._asyncRequests["price"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["price"];
      let updateStateObj = { isLoading: false };
      if (!resp.error) {
        updateStateObj.price = resp.price
        this.props.onNewPrice(resp.price);
      }

      _this.setState(updateStateObj);
    });
  }

  componentWillUnmount() {
    if (this._asyncRequests["price"]) {
      delete this._asyncRequests["price"];
    }
  }

  render() {
    let title = (<PriceFeedTitle
      label={this.props.title}
      loading={false}
      price={this.state.price}
      fetchPrice={this.queryPrice}
      />);
    if (this.state.isLoading || !this.props.optionsMetadata) {
      title = (<PriceFeedTitle
        label={this.props.title}
        loading={true}
        fetchPrice={this.queryPrice}
        />);
    }

    let values = [this.props.type];
    if (this.props.type === "exchange") {
      let parts = this.props.feed_url.split('/');
      values.push(parts[0]);
      if (parts.length > 1) {
        // then it has to be 3
        values.push(parts[1] + "/" + parts[2]);
      }
    } else {
      values.push(this.props.feed_url);
    }

    let selector = (<PriceFeedSelector
      optionsMetadata={this.props.optionsMetadata}
      values={values}
      onChange={this.props.onChange}
      />
    );
    if (!this.props.optionsMetadata) {
      selector = (<div className={styles.loaderWrapper}>
          <LoadingAnimation/>
        </div>
      );
    }

    return (
      <div>
        {title}
        {selector}
      </div>
    );
  }
}

export default PriceFeedAsset;