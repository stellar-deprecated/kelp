import React, { Component } from 'react';
import PropTypes from 'prop-types';
import grid from '../../_styles/grid.module.scss';
import Button from '../../atoms/Button/Button';
import Label from '../../atoms/Label/Label';

class SecretKey extends Component {
  static propTypes = {
    label: PropTypes.string.isRequired,
    isTestNet: PropTypes.bool.isRequired,
    secret: PropTypes.element.isRequired,
    onNewKeyClick: PropTypes.func,
    optional: PropTypes.bool,
  };

  render() {
    let secret = this.props.secret;
    if (this.props.isTestNet) {
      secret = (
        <div className={grid.row}>
          <div className={grid.col90p}>
            {this.props.secret}
          </div>
          <div className={grid.col10p}>
            <Button 
              icon="refresh"
              size="small"
              hsize="round"
              loading={false}
              onClick={this.props.onNewKeyClick}
              />
          </div>
        </div>
      );
    }

    let label = (<Label>{this.props.label}</Label>);
    if (this.props.optional) {
      label = (<Label optional>{this.props.label}</Label>);
    }

    return (
      <div>
        {label}
        {secret}
      </div>
    );
  }
}

export default SecretKey;