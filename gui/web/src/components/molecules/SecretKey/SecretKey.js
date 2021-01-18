import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './SecretKey.module.scss';
import StellarSdk from 'stellar-sdk'
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
    readOnly: PropTypes.bool,
    eventPrefix: PropTypes.string.isRequired,
  };

  render() {
    let inputElem = (<Input
      value={this.props.secret}
      type="string"
      onChange={(event) => { this.props.onSecretChange(event) }}
      error={this.props.onError()}
      readOnly={this.props.readOnly}
      />);

    let secretElem = inputElem;
    if (this.props.isTestNet && !this.props.readOnly) {
      secretElem = (
        <div className={grid.row}>
          <div className={grid.col90p}>
            {inputElem}
          </div>
          <div className={grid.col10p}>
            <Button
              eventName={this.props.eventPrefix + "-new"}
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

    let pubkeyElem = ""
    if (this.props.secret !== "") {
      let pubkey = "";
      try {
        let pubkeypair = StellarSdk.Keypair.fromSecret(this.props.secret);
        pubkey = pubkeypair.publicKey();
      } catch (error) {
        pubkey = "<" + error.message + ">";
      }

      pubkeyElem = (
        <div>
          <span className={styles.pubkeyLabel}>PubKey: </span><span className={styles.pubkey}>{pubkey}</span>
        </div>
      );
    }

    return (
      <div>
        {label}
        {pubkeyElem}
        {secretElem}
      </div>
    );
  }
}

export default SecretKey;