import React, { Component } from 'react';
import styles from './Icon.module.scss';

import day from '../../../assets/images/ico-day.svg';
import help from '../../../assets/images/ico-help.svg';


const viewBox = {
  day: '0 0 19 19',
  help: '0 0 19 19',
}

const icons = {
  day,
  help
}

class Icon extends Component {
  static defaultProps = {
    width: '22px',
    height: '22px',
  }
  
  render() {
    return (
      <div>
        <svg 
          viewBox={viewBox[this.props.icon]}
          width={this.props.width}
          height={this.props.height}
        >
          <use xlinkHref={icons[this.props.icon]+'#ico'}/>
        </svg>
      </div>  
    );
  }
}

export default Icon;