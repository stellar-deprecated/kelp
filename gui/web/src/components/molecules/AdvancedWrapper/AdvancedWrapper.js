import React, { Component } from 'react';
import classNames from 'classnames';
import styles from './AdvancedWrapper.module.scss';
import Icon from '../../atoms/Icon/Icon';

class AdvancedWrapper extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isOpened: props.isOpened,
    };
    this.toggleOpen = this.toggleOpen.bind(this);
  }

  toggleOpen() {
    if(this.state.isOpened){
      this.close();
    }
    else {
      this.open();
    }
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

    let classNameList = classNames(
      styles.wrapper,
      isOpenedClass
    );

    return (
      <div className={classNameList}>
        <div className={this.props.headerClass}>
          <div className={styles.header} onClick={this.toggleOpen}>
            <Icon symbol="caretDown" width="12px" height="12px"/>
            Advanced Settings
          </div>
        </div>
        <div className={styles.content}>
          <div className={styles.contentWrapper}>
            {this.props.children}
          </div>
        </div>
      </div>
    );
  }
}

export default AdvancedWrapper;