import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import styles from './Button.module.scss';
import Icon from '../Icon/Icon';
import LoadingAnimation from '../LoadingAnimation/LoadingAnimation';
import sendMetricEvent from '../../../kelp-ops-api/sendMetricEvent';

const iconSizes = {
  small:'10px',
  medium: '13px',
  large: '16px',
}

const iconSizesRound = {
  small:'12px',
  medium: '14px',
  large: '19px',
}

class Button extends Component {
  constructor(props) {
    super(props);
    this.trackOnClick = this.trackOnClick.bind(this);
    this.sendMetricEvent = this.sendMetricEvent.bind(this);

    this._asyncRequests = {};
  }

  static defaultProps = {
    icon: null,
    size: 'medium',
    variant: '',
    hsize: 'regular',
    loading: false,
    onClick: () => {},
    disabled: false    
  }

  static propTypes = {
    icon: PropTypes.string,
    size: PropTypes.string,
    variant: PropTypes.string,
    onClick: PropTypes.func,
    loading: PropTypes.bool,
    disabled: PropTypes.bool,
    eventName: function(props, propName, componentName) {
      if (!/^[a-zA-Z0-9]+$/.test(props[propName])) {
        return new Error('Invalid prop `' + propName + '` supplied to `' + componentName + '`. Validation failed.');
      }
    },
  };

  sendMetricEvent() {
    if (this._asyncRequests["sendMetricEvent"]) {
      return
    }

    if (this.props.eventName === "" || this.props.eventName === undefined) {
      return
    }

    const _this = this;
    const eventData = {
      eventName: this.props.eventName,
      type: "generic",
      component: "button"
    };
    this._asyncRequests["sendMetricEvent"] = sendMetricEvent(this.props.baseUrl, this.props.eventName, eventData).then(resp => {
      if (!_this._asyncRequests["sendMetricEvent"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }
      delete _this._asyncRequests["sendMetricEvent"];

      if (!resp.success) {
        console.log(resp.error);
      }
      // else do nothing on success
    });
  }

  trackOnClick() {
    this.sendMetricEvent();
    this.props.onClick();
  }

  render() {
    const iconOnly = this.props.children ? null : styles.iconOnly;
    const isLoading = this.props.loading ? styles.isLoading : null;
    const iconSize = this.props.hsize === 'round' ? iconSizesRound[this.props.size] : iconSizes[this.props.size];

    const classNameList = classNames(
      this.props.className,
      styles.button, 
      styles[this.props.size],
      styles[this.props.hsize],
      styles[this.props.variant],
      iconOnly,
      isLoading,

    );

    return (
      <button 
        className={classNameList} 
        disabled={this.props.disabled || this.props.loading } 
        onClick= {this.trackOnClick}
      >
        {this.props.loading &&
          <span className={styles.loader}>
            <LoadingAnimation/> 
          </span>
        }

        <span className={styles.content}>          
          { this.props.icon && (
            <Icon 
              symbol={this.props.icon} 
              width={iconSize}
              height={iconSize}
            />
          )}
          {this.props.children}
        </span>
      </button>
    );
  }
}

export default Button;