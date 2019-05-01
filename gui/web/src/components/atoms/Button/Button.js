import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import styles from './Button.module.scss';
import Icon from '../Icon/Icon';

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
  static defaultProps = {
    icon: null,
    size: 'medium',
    variant: '',
    hsize: 'regular',
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

    const iconSize = this.props.hsize == 'round' ? iconSizesRound[this.props.size] : iconSizes[this.props.size];
    console.log(this.props.hsize);
    console.log(iconSize);

    const classNameList = classNames(
      this.props.className,
      styles.button, 
      styles[this.props.size],
      styles[this.props.hsize],
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
              width={iconSize}
              height={iconSize}
            />
          )}
          {this.props.children}
        </button>
    );
  }
}

export default Button;