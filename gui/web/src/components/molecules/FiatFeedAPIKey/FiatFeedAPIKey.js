import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Input from '../../atoms/Input/Input';
import Label from '../../atoms/Label/Label';

class FiatFeedAPIKey extends Component {
    static propTypes = {
        enabled: PropTypes.bool,
        value: PropTypes.string.isRequired,
        onChange: PropTypes.func,
        error: PropTypes.string,
    };

    render() {
        if (!this.props.enabled) {
            return "";
        }

        return (
            <div>
                <Label>Fiat API Key (Currency Layer)</Label>
                <Input
                    value={this.props.value}
                    type="string"
                    onChange={this.props.onChange}
                    error={this.props.error}
                    />
            </div>
        );
    }
}

export default FiatFeedAPIKey;