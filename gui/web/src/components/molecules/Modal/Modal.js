import React, { Component } from 'react';
import PropTypes from 'prop-types';
import styles from './Modal.module.scss';
import classNames from 'classnames';
import Icon from '../../atoms/Icon/Icon';
import Button from '../../atoms/Button/Button';

class Modal extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpened: true,
    };
    this.close = this.close.bind(this);
    this.action = this.action.bind(this);
  }
  
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
  };

  open() {
    this.setState({
      isOpened: true,
    })
  }

  close() {
    this.setState({
      isOpened: false,
    })
    this.props.onClose();
  }

  action() {
    this.setState({
      isOpened: false,
    })
    this.props.onAction();
  }

  render() {
    let isOpenedClass = this.state.isOpened ? styles.isOpened : null;

    let wrapperClasses = classNames(
      styles.wrapper,
      isOpenedClass,
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

    const actionButton = (
      <Button onClick={this.action}>
        {this.props.actionLabel}
      </Button>
    );

    return (
      <div className={wrapperClasses}>
        <div className={styles.window}>
          <Button 
            icon="close"
            size="small"
            variant="transparent"
            hsize="round"
            className={styles.closeButton} 
            onClick={this.close}
          />
          {iconTag}
          {titleTag}
          {textTag}
          {bulletsTag}
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