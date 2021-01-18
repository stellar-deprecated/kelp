import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Modal.module.scss';
import classNames from 'classnames';
import Icon from '../../atoms/Icon/Icon';
import Button from '../../atoms/Button/Button';

class Modal extends Component {
  static defaultProps = {
    text: null,
    bullets: [],
    actionLabel: 'Close',
  }

  static propTypes = {
    type: PropTypes.string.isRequired,
    title: PropTypes.string.isRequired,
    onClose: PropTypes.func.isRequired,
    text: PropTypes.string,
    bullets: PropTypes.array,
    actionLabel: PropTypes.string,
    onAction: PropTypes.func,
    onPrevious: PropTypes.func, // displays button if the function is defined
    onNext: PropTypes.func, // displays button if the function is defined
  };

  render() {
    let wrapperClasses = classNames(
      styles.wrapper,
      styles.isOpened,
    );

    const bulletsClasses = classNames(
      styles.bullets,
      styles[this.props.type],
    );

    let iconTag = null;
    if (this.props.type) {
      iconTag = (
        <Icon 
        symbol={this.props.type} 
        width="50px" 
        height="50px"
      />);
    }

    let titleTag = null;
    if (this.props.title) {
      titleTag = (<h3 className={styles.title}>{this.props.title}</h3>);
    }

    let textTag = null;
    if (this.props.text) {
      textTag = (<p className={styles.text}>{this.props.text}</p>);
    }

    let bulletsTag = null;
    if (this.props.bullets.length > 0) {
      const liList = this.props.bullets.map((item, index) => (
        <li key={index}>{item}</li>
       ));
      bulletsTag = (
        <ul className={bulletsClasses}>
         {liList}
        </ul>
      );
    }

    const eventPrefix = "modal-errors";
    let prevButton = null;
    if (this.props.onPrevious) {
      prevButton = (<Button eventName={eventPrefix + "-previous"} onClick={this.props.onPrevious} icon="caretLeft" size="medium" className={styles.arrowButton}/>);
    }
    const actionButton = (
      <Button eventName={eventPrefix + "-action"} onClick={this.props.onAction}>
        {this.props.actionLabel}
      </Button>
    );
    let nextButton = null;
    if (this.props.onNext) {
      nextButton = (<Button eventName={eventPrefix + "-next"} onClick={this.props.onNext} icon="caretRight" size="medium" className={styles.arrowButton}/>);
    }

    return (
      <div className={wrapperClasses}>
        <div className={styles.window}>
          <Button
            eventName={eventPrefix + "-close"}
            icon="close"
            size="small"
            variant="transparent"
            hsize="round"
            className={styles.closeButton} 
            onClick={this.props.onClose}
          />
          {iconTag}
          {titleTag}
          {textTag}
          <div className={styles.middleContainer}>
            <div className={styles.prevButtonContainer}>{prevButton}</div>
            <div className={styles.bulletsContainer}>{bulletsTag}</div>
            <div className={styles.nextButtonContainer}>{nextButton}</div>
          </div>
          <div className={styles.footer}>
            {actionButton}
          </div>
        </div>
        <span className={styles.backdrop}/>
      </div>
    );
  }
}

export default Modal;