import React, { Component } from 'react';
import PropTypes from 'prop-types';
import grid from '../../_styles/grid.module.scss';
import Button from '../../atoms/Button/Button';
import Label from '../../atoms/Label/Label';
import Input from '../../atoms/Input/Input';

class SecretKey extends Component {
  static propTypes = {
    label: PropTypes.string.isRequired,
    isTestNet: PropTypes.bool.isRequired,
    secret: PropTypes.string.isRequired,
    onSecretChange: PropTypes.func.isRequired,
    onError: PropTypes.func.isRequired,
    onNewKeyClick: PropTypes.func,
    optional: PropTypes.bool,
  };

  render() {
    let inputElem = (<Input
      value={this.props.secret}
      type="string"
      onChange={(event) => { this.props.onSecretChange(event) }}
      error={this.props.onError}
      />);

    let secretElem = inputElem;
    if (this.props.isTestNet) {
      secretElem = (
        <div className={grid.row}>
          <div className={grid.col90p}>
            {inputElem}
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
        {secretElem}
      </div>
    );
  }
}

export default SecretKey;