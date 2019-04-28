import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Button.module.scss';
import classNames from 'classnames';



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
    const className = classNames(
      styles.button, 
      styles[this.props.size],
      styles[this.props.variant],
    );

    return (
        <button className={className} disabled={this.props.disabled} onClick={this.props.onClick}>
          {this.props.children}
        </button>
    );
  }
}

export default Button;