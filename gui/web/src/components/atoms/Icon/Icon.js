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
import close from '../../../assets/images/ico-close.svg';
import warning from '../../../assets/images/ico-alert.svg';
import error from '../../../assets/images/ico-error.svg';
import refresh from '../../../assets/images/ico-refresh.svg';
import remove from '../../../assets/images/ico-remove.svg';
import options from '../../../assets/images/ico-options.svg';
import back from '../../../assets/images/ico-arrow-back.svg';


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
  close: '0 0 12 12',
  error: '0 0 50 50',
  warning: '0 0 50 50',
  refresh: '0 0 15 15',
  remove: '0 0 11 11',
  options: '0 0 20 20',
  back: '0 0 14 14',
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
  arrowDown,
  close,
  error,
  warning,
  refresh,
  remove,
  options,
  back,
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
        viewBox={viewBox[this.props.symbol]}
        width={this.props.width}
        height={this.props.height}
      >
        <use xlinkHref={icons[this.props.symbol]+'#ico'}/>
      </svg>
    );
  }
}

export default Icon;