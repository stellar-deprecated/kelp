import React, { Component } from 'react';
import styles from './SectionTitle.module.scss';

class SectionTitle extends Component {

  render() {
    return (
      <h3 className={styles.title}>{this.props.children}</h3>
    );
  }
}

export default SectionTitle;