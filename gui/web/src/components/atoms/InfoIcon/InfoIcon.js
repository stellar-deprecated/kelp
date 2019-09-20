import React, { Component } from 'react';
import styles from './InfoIcon.module.scss';
import PropTypes from 'prop-types';
import Tooltip from 'react-tooltip-lite';
import Icon from '../Icon/Icon';

class InfoIcon extends Component {
  static propTypes = {
    issuer: PropTypes.string,
  }

  render() {
    if (!this.props.issuer || this.props.issuer === "") {
      return "";
    }

    const tip = (<div className={styles.tooltip}>{this.props.issuer}</div>);
    return (
      <div className={styles.wrapper}>
        <Tooltip
          content={tip}
          tipContentClassName={styles.tooltipWrapper}
          useHover={true}
          hoverDelay={50}
          arrowSize={5}
          padding={"0px"}
          tagName="div"
          tipContentHover={true}
          >
          <Icon className={styles.icon} symbol="info" width="8" height="8"/>
        </Tooltip>
      </div>
    );
  }
}

export default InfoIcon;