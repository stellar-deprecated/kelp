import React, { Component } from 'react';
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
  }
  
  static defaultProps = {
    type: null,
    text: null,
    bullets: [],
    actionLabel: 'Close',
  }

  open() {
    this.setState({
      isOpened: true,
    })
  }

  close() {
    this.setState({
      isOpened: false,
    })
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

          {this.props.type && (
            <Icon 
            symbol={this.props.type} 
            width="50px" 
            height="50px"
          />
          )}
          
          {this.props.title && (
            <h3 className={styles.title}>{this.props.title}</h3>
          )}

          {this.props.text && (
          <p className={styles.text}>{this.props.text}</p>
          )}

          {this.props.bullets.length > 0 && (
            <ul className={bulletsClasses}>
             {this.props.bullets.map((item, index) => (
              <li key={index}>{item}</li>
             ))}
            </ul>
          )}

          <div className={styles.footer}>
            <Button onClick={this.close}>{this.props.actionLabel}</Button>
          </div>
        </div>
        <span className={styles.backdrop}/>
      </div>
    );
  }
}

export default Modal;