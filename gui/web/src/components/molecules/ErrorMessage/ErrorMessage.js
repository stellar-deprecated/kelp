import React, { Component } from 'react';
import styles from './ErrorMessage.module.scss';

class ErrorMessage extends Component {
  render() {

    return (
      <div className={styles.wrapper}>
        <p className={styles.title}>Oops, something is not right.
        </p>
        <p className={styles.text}>Please, review the fields marked in red and try again.
        </p>
      </div>
    );
  }
}

export default ErrorMessage;