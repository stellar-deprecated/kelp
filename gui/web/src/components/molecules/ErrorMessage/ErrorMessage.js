import React, { Component } from 'react';
import styles from './ErrorMessage.module.scss';
import PropTypes from 'prop-types';

class ErrorMessage extends Component {
  static propTypes = {
    errorList: PropTypes.arrayOf(PropTypes.string),
  };

  render() {
    let errorElem = "";
    if (this.props.errorList.length > 1) {
      let errorItems = [];
      for (let i in this.props.errorList) {
        let e = this.props.errorList[i];
        errorItems.push(<li className={styles.text} key={e}>{e}</li>)
      }
      errorElem = (<ul>{errorItems}</ul>);
    } else {
      errorElem = (<p className={styles.text}>{this.props.errorList[0]}</p>);
    }

    return (
      <div className={styles.wrapper}>
        <p className={styles.title}>Oops, something is not right</p>
        {errorElem}
      </div>
    );
  }
}

export default ErrorMessage;