import React, { Component } from 'react';
import styles from'./RemoveButton.module.scss';
import Button from '../Button/Button';


class RemoveButton extends Component {
  render() {
    return (
      <Button 
        icon="remove" 
        variant="danger" 
        className={styles.button}
        hsize="round">
      </Button>
    );
  }
}

export default RemoveButton;