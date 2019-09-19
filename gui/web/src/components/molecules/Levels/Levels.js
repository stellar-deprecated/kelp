import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Levels.module.scss';
import Button from '../../atoms/Button/Button';
import FieldItem from '../FieldItem/FieldItem';
import Label from '../../atoms/Label/Label';
import Input from '../../atoms/Input/Input';
import ErrorMessage from '../ErrorMessage/ErrorMessage';

class Levels extends Component {
  static propTypes = {
    levels: PropTypes.arrayOf(PropTypes.shape({
      spread: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
      amount: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
    })).isRequired,
    updateLevel: PropTypes.func.isRequired,
    newLevel: PropTypes.func.isRequired,
    hasNewLevel: PropTypes.func.isRequired,
    onRemove: PropTypes.func.isRequired,
    levelErrors: PropTypes.object,
    addLevelError: PropTypes.func.isRequired,
    clearLevelError: PropTypes.func.isRequired
  }

  render() {
    let levels = [];
    for (let i = 0; i < this.props.levels.length; i++) {
      let levelIdx = "" + i;
      let levelError = this.props.levelErrors[levelIdx];
      if (!levelError) {
        levelError = {
          spread: null,
          amount: null,
        };
      }
      
      levels.push((
        <div key={i} className={styles.item}>
          <div>
            <FieldItem>
              <Label>Spread</Label>
              <Input
                suffix="%"
                value={this.props.levels[i].spread}
                type="percent_positive"
                invokeChangeOnLoad={true}
                onChange={(event) => { this.props.updateLevel(i, "spread", event.target.value) }}
                error={levelError.spread}
                triggerError={(message) => { this.props.addLevelError(i, "spread", message) }}
                clearError={() => { this.props.clearLevelError(i, "spread") }}
                />
            </FieldItem>
            <FieldItem>
              <Label>Amount</Label>
              <Input
                value={this.props.levels[i].amount}
                type="float_positive"
                invokeChangeOnLoad={true}
                onChange={(event) => { this.props.updateLevel(i, "amount", event.target.value) }}
                error={levelError.amount}
                triggerError={(message) => { this.props.addLevelError(i, "amount", message) }}
                clearError={() => { this.props.clearLevelError(i, "amount") }}
                />
            </FieldItem>
          </div>
          <div className={styles.actions}>
            <Button 
              className={styles.button}
              icon="remove" 
              variant="danger" 
              onClick={() => this.props.onRemove(i)}
              hsize="round"
              />
          </div>
        </div>
      ));
    }

    let error = "";
    if (this.props.error) {
      error = (<ErrorMessage errorList={[this.props.error]}/>);
    }

    return (
      <div className={styles.wrapper}>
        {levels}
        {error}
        <Button
          className={styles.add}
          icon="add"
          variant="faded"
          onClick={this.props.newLevel}
          disabled={this.props.hasNewLevel()}
          >
            New Level
          </Button>
      </div>
    );
  }
}

export default Levels;