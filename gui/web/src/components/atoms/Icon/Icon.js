import React, { Component } from 'react';
import styles from './Icon.module.scss';

import day from '../../../assets/images/ico-day.svg';
import help from '../../../assets/images/ico-help.svg';
import info from '../../../assets/images/ico-info.svg';
import warningSmall from '../../../assets/images/ico-warning-small.svg';
import errorSmall from '../../../assets/images/ico-error-small.svg';
import add from '../../../assets/images/ico-add.svg';
import download from '../../../assets/images/ico-download.svg';
import start from '../../../assets/images/ico-start.svg';
import stop from '../../../assets/images/ico-stop.svg';
import wave from '../../../assets/images/ico-wave.svg';
import arrowDown from '../../../assets/images/ico-arrow-down.svg';


const viewBox = {
  day: '0 0 19 19',
  help: '0 0 19 19',
  warningSmall: '0 0 11 11',
  errorSmall: '0 0 11 11',
  info: '0 0 8 8',
  add: '0 0 13 13',
  download: '0 0 13 13',
  start: '0 0 8 8',
  stop: '0 0 8 8',
  wave: '0 0 20 4',
  arrowDown: '0 0 12 7',
}

const icons = {
  day,
  help,
  warningSmall,
  errorSmall,
  info,
  add,
  download,
  start,
  stop,
  wave,
  arrowDown
}

class Icon extends Component {
  static defaultProps = {
    symbol: null,
    width: '22px',
    height: '22px',
  }
  
  render() {
    return (
      <svg 
        className={this.props.className}
        viewBox={viewBox[this.props.icon]}
        width={this.props.width}
        height={this.props.height}
      >
        <use xlinkHref={icons[this.props.symbol]+'#ico'}/>
      </svg>
    );
  }
}

export default Icon;