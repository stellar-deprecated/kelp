import React, { Component } from 'react';
import styles from './RepeaterField.module.scss';
import RemoveButton from '../../atoms/RemoveButton/RemoveButton';
import Button from '../../atoms/Button/Button';

class RepeaterField extends Component {
  static defaultProps = {
    groupTitle: "",
  }

  render() {

    return (
      <div className={styles.wrapper}>
        <div className={styles.item}>
          <div>
            {this.props.children}
          </div>
          <div className={styles.actions}>
            <RemoveButton/>
          </div>
        </div>
        <div className={styles.item}>
          <div>
            {this.props.children}
          </div>
          <div className={styles.actions}>
            <RemoveButton/>
          </div>
        </div>
        <div className={styles.item}>
          <div>
            {this.props.children}
          </div>
          <div className={styles.actions}>
            <RemoveButton/>
          </div>
        </div>
        <Button icon="add" variant="faded" className={styles.add}>New Level</Button>
      </div>
    );
  }
}

export default RepeaterField;