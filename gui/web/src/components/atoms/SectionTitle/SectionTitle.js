import React, { Component } from 'react';
import styles from './SectionTitle.module.scss';

class SectionTitle extends Component {

  render() {
    return (
      <div className={this.props.className}>
        <h3 className={styles.title}>{this.props.children}</h3>
      </div>
    );
  }
}

export default SectionTitle;