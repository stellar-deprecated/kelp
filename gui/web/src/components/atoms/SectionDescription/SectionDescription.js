import React, { Component } from 'react';
import styles from './SectionDescription.module.scss';

class SectionDescription extends Component {

  render() {
    return (
      <p className={styles.text}>{this.props.children}</p>
    );
  }
}

export default SectionDescription;