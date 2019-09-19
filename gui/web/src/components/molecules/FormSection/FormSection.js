import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './FormSection.module.scss';
import classNames from 'classnames';
import grid from '../../../components/_styles/grid.module.scss';

class FormSection extends Component {
  static propTypes = {
    wideCol80: PropTypes.bool,
    wideCol100: PropTypes.bool,
    tip: PropTypes.string,
  };
  
  render() {
    let tipWrapperClasses = classNames(
      styles.tipWrapper,
      grid.col5,
    );

    let colClassName = grid.col7;
    if (this.props.wideCol80) {
      colClassName = grid.col80p;
    } else if (this.props.wideCol100) {
      colClassName = grid.col100p;
    }

    return (
      <div className={grid.row}>
        <div className={colClassName}>
          {this.props.children}
        </div>
        {this.props.tip && (
          <div className={tipWrapperClasses}>
            <div className={styles.tip}>
              {this.props.tip}
            </div>  
          </div>
        )}
      </div>
    );
  }
}

export default FormSection;