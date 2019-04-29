import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Button.module.scss';
import classNames from 'classnames';
import Icon from '../Icon/Icon';

const iconSizes = {
  small:'10px',
  medium: '13px',
  large: '16px',
}

class Button extends Component {
  static defaultProps = {
    icon: null,
    size: 'medium',
    variant: 'regular',
    onClick: () => {},
    disabled: false    
  }

  static propTypes = {
    icon: PropTypes.string,
    size: PropTypes.string,
    variant: PropTypes.string,
    onClick: PropTypes.func,
    disabled: PropTypes.bool
  };

  render() {
    const iconOnly = this.props.children ? null : styles.iconOnly;

    const classNameList = classNames(
      styles.button, 
      styles[this.props.size],
      styles[this.props.variant],
      iconOnly,
    );

    return (
        <button 
          className={classNameList} 
          disabled={this.props.disabled} 
          onClick= {this.props.onClick}
        >
          { this.props.icon && (
            <Icon 
              symbol={this.props.icon} 
              width={iconSizes[this.props.size]}
              height={iconSizes[this.props.size]}
            />
          )}
          {this.props.children}
        </button>
    );
  }
}

export default Button;