import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import Constants from '../../../Constants';
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
    // we specify a custom validator. It should return an Error object if the validation fails
    // don't `console.warn` or throw, as this won't work inside `oneOfType`.
    eventName: function(props, propName, componentName) {
      // "-" needs to be first or last character to be used literally
      // source: https://stackoverflow.com/questions/8833963/allow-dash-in-regular-expression
      if (!/^[-a-zA-Z0-9]+$/.test(props[propName])) {
        return new Error('Invalid prop `' + propName + '` supplied to `' + componentName + '`. Validation failed.');
      }
    },
  };

  sendMetricEvent() {
    if (this._asyncRequests["sendMetricEvent"]) {
      return
    }

    if (!this.props.eventName || this.props.eventName === "") {
      console.error("programmer error: no event name provided for this Button, not sending button click event!");
      return
    }

    const _this = this;
    const eventData = {
      gui_event_name: this.props.eventName,
      gui_category: "generic",
      gui_component: "button"
    };
    this._asyncRequests["sendMetricEvent"] = sendMetricEvent(Constants.BaseURL, "gui-button", eventData).then(resp => {
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