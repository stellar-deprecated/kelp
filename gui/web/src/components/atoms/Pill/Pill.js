import React, { Component } from 'react';
import styles from './Pill.module.scss';
import Icon from '../Icon/Icon';




class Pill extends Component {
  static defaultProps = {
    type: null,
    number: 0,
  }

  render() {
    if(this.props.number){
      let symbolName = null
      if(this.props.type === 'warning'){
        symbolName = 'warningSmall'
      }
      else {
        symbolName = 'errorSmall'
      }

      return (
          <div className={styles[this.props.type]}>
            <Icon className={styles.icon} symbol={symbolName} width={'11px'} height={'11px'}></Icon>
            <span>
              {this.props.number}
            </span>
          </div>
      );
    }
    else {
      return null;
    }
  }
}

export default Pill;