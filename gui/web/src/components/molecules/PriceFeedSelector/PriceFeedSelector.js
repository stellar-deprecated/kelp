import React, { Component } from 'react';
import styles from './PriceFeedSelector.module.scss';
import Select from '../../atoms/Select/Select';
import grid from '../../../components/_styles/grid.module.scss';

class PriceFeedSelector extends Component {
  static defaultProps = {
    groupTitle: "",
  }

  render() {

    return (
      <div className={styles.wrapper}>
        <div className={grid.row}>
          <div className={grid.col6}>
            <Select/>
          </div>
          <div className={grid.col3}>
            <Select/>
          </div>
          <div className={grid.col3}>
            <Select/>
          </div>
        </div>
      </div>
    );
  }
}

export default PriceFeedSelector;