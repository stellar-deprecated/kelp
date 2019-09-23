import React, { Component } from 'react';
import styles from './FiatFeedAPIKey.module.scss';
import PropTypes from 'prop-types';
import Input from '../../atoms/Input/Input';
import Label from '../../atoms/Label/Label';

class FiatFeedAPIKey extends Component {
    static propTypes = {
        enabled: PropTypes.bool,
        value: PropTypes.string.isRequired,
        onChange: PropTypes.func,
        error: PropTypes.string,
        readOnly: PropTypes.bool,
    };

    render() {
        return (
            <div className={styles.apiKey}>
                <Label disabled={!this.props.enabled}>Fiat API Key (Currency Layer)</Label>
                <Input
                    value={this.props.value}
                    type="string"
                    onChange={this.props.onChange}
                    error={this.props.error}
                    disabled={!this.props.enabled}
                    readOnly={this.props.readOnly}
                    />
            </div>
        );
    }
}

export default FiatFeedAPIKey;