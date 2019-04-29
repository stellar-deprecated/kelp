import React, { Component } from 'react';
import styles from './Modal.module.scss';
import Icon from '../../atoms/Icon/Icon';
import Button from '../../atoms/Button/Button';

class Modal extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpened: false,
    };
    // this.open = this.toggleOpen.bind(this);
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
    return (
      <div className={classNameList}>
        <div className={styles.window}>
          <Button className={styles.closeButton}/>
          <Icon />
          <h3 className={styles.title}>{this.props.title}</h3>
          <p className={styles.text}>{this.props.text}</p>
          <ul className={styles.text}>
            <li>item1</li>
            <li>item2</li>
            <li>item3</li>
          </ul>
          <div className={styles.footer}>
            <Button>Go to bot settings</Button>
          </div>
        </div>
      </div>
    );
  }
}

export default Modal;