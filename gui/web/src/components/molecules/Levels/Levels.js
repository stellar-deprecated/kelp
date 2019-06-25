import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Levels.module.scss';
import Button from '../../atoms/Button/Button';
import FieldItem from '../FieldItem/FieldItem';
import Label from '../../atoms/Label/Label';
import Input from '../../atoms/Input/Input';

class Levels extends Component {
  static defaultProps = {
    levels: PropTypes.arrayOf(PropTypes.shape({
      spread: PropTypes.string,
      amount: PropTypes.string,
    })).isRequired,
    updateLevel: PropTypes.func.isRequired,
    newLevel: PropTypes.func.isRequired,
    hasNewLevel: PropTypes.func.isRequired,
    onRemove: PropTypes.func.isRequired
  }

  render() {
    let levels = [];
    for (let i = 0; i < this.props.levels.length; i++) {
      levels.push((
        <div key={i} className={styles.item}>
          <div>
            <FieldItem>
              <Label>Spread</Label>
              <Input
                suffix="%"
                value={this.props.levels[i].spread}
                type="percent"
                onChange={(event) => { this.props.updateLevel(i, "spread", event.target.value) }}
                />
            </FieldItem>
            <FieldItem>
              <Label>Amount</Label>
              <Input
                value={this.props.levels[i].amount}
                type="float"
                onChange={(event) => { this.props.updateLevel(i, "amount", event.target.value) }}
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

    return (
      <div className={styles.wrapper}>
        {levels}
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