import React, { Component } from 'react';
import styles from './FormSection.module.scss';
import classNames from 'classnames';
import grid from '../../../components/_styles/grid.module.scss';

class FormSection extends Component {
  
  render() {
    let tipWrapperClasses = classNames(
      styles.tipWrapper,
      grid.col5,
    );

    return (
      <div className={grid.row}>
        <div className={grid.col7}>
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