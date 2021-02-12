import React, { Component } from 'react';
import PropTypes from 'prop-types';
import BotCard from '../../molecules/BotCard/BotCard';
import Button from '../../atoms/Button/Button';
import EmptyList from '../../molecules/EmptyList/EmptyList';
import ScreenHeader from '../../molecules/ScreenHeader/ScreenHeader';
import grid from '../../_styles/grid.module.scss';
import autogenerate from '../../../kelp-ops-api/autogenerate';
import listBots from '../../../kelp-ops-api/listBots';
import Constants from '../../../Constants';
import Modal from '../../molecules/Modal/Modal';

class Bots extends Component {
  constructor(props) {
    super(props);
    this.state = {
      bots: [],
    };
 
    this.fetchBots = this.fetchBots.bind(this);
    this.gotoDetails = this.gotoDetails.bind(this);
    this.autogenerateBot = this.autogenerateBot.bind(this);
    this.createBot = this.createBot.bind(this);
    
    this._asyncRequests = {};
  }

  static propTypes = {
    baseUrl: PropTypes.string.isRequired,
    activeError: PropTypes.object,  // can be null
    setActiveError: PropTypes.func.isRequired,  // (botName, level, errorList, index)
    addError: PropTypes.func.isRequired,  // (backendError)
    removeError: PropTypes.func.isRequired,  // (object_name, level, error)
    hideActiveError: PropTypes.func.isRequired, // ()
    findErrors: PropTypes.func.isRequired, // (object_name, level)
    enablePubnetBots: PropTypes.bool.isRequired,
  };

  componentWillUnmount() {
    if (this._asyncRequests["listBots"]) {
      delete this._asyncRequests["listBots"];
    }

    if (this._asyncRequests["autogenerate"]) {
      delete this._asyncRequests["autogenerate"];
    }
  }

  componentDidMount() {
    this.fetchBots()
  }

  gotoDetails(botName) {
    this.props.history.push('/details?bot_name=' + encodeURIComponent(botName))
  }

  fetchBots() {
    if (this._asyncRequests["listBots"]) {
      return
    }

    var _this = this
    this._asyncRequests["listBots"] = listBots(this.props.baseUrl).then(bots => {
      if (!_this._asyncRequests["listBots"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["listBots"];
      if (bots.hasOwnProperty('error')) {
        console.log("error in listBots: " + bots.error);
        setTimeout(this.fetchBots, 1000);
      } else {
        _this.setState(prevState => ({
          bots: bots
        }))
      }
    });
  }

  autogenerateBot() {
    var _this = this
    this._asyncRequests["autogenerate"] = autogenerate(this.props.baseUrl).then(newBot => {
      if (!_this._asyncRequests["autogenerate"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["autogenerate"];
      _this.setState(prevState => ({
        bots: [...prevState.bots, newBot]
      }))
    });
  }
  
  createBot() {
    this.props.history.push('/new')
  }

  render() {
    let inner = <EmptyList autogenerateFn={this.autogenerateBot} createBotFn={this.createBot}/>;
    if (this.state.bots.length) {
      let screenHeader = (
        <ScreenHeader title={'My Bots'}>
          <Button 
            eventName={"main-newbot"}
            variant="faded" 
            hsize="short"
            icon="add" 
            onClick={this.createBot}
            >
              New Bot
            </Button>
        </ScreenHeader>
      );

      let cards = this.state.bots.map((bot, index) => {
        const errorLevelInfoForBot = this.props.findErrors(bot.name, Constants.ErrorLevel.info);
        const errorLevelWarningForBot = this.props.findErrors(bot.name, Constants.ErrorLevel.warning);
        const errorLevelErrorForBot = this.props.findErrors(bot.name, Constants.ErrorLevel.error);

        return <BotCard
          key={index} 
          name={bot.name}
          enablePubnetBots={this.props.enablePubnetBots}
          history={this.props.history}
          running={bot.running}
          addError={(kelpError) => this.props.addError(kelpError)}
          errorLevelInfoForBot={errorLevelInfoForBot}
          errorLevelWarningForBot={errorLevelWarningForBot}
          errorLevelErrorForBot={errorLevelErrorForBot}
          setModal={(level, errorList) => {
            // index is always 0 here because incrementing the index happens in App.js when we traverse the errorList, never when we open the modal for the first time
            this.props.setActiveError(bot.name, level, errorList, 0);
          } }
          // showDetailsFn={this.gotoDetails}
          baseUrl={this.props.baseUrl}
          reload={this.fetchBots}
        />
      });

      inner = (
        <div>
          {screenHeader}
          {cards}
        </div>
      );
    } else {
      setTimeout(this.fetchBots, 1000);
    }

    const activeError = this.props.activeError;
    let modalWindow = null;
    if (activeError) {
      const indexedError = activeError.errorList[activeError.index];
      let onPrevious = null;
      if (activeError.index > 0) {
        onPrevious = () => {this.props.setActiveError(activeError.botName, activeError.level, activeError.errorList, activeError.index - 1)};
      }
      let onNext = null;
      if (activeError.index < activeError.errorList.length - 1) {
        onNext = () => {this.props.setActiveError(activeError.botName, activeError.level, activeError.errorList, activeError.index + 1)};
      }
      modalWindow = (<Modal
        type={activeError.level}
        title={indexedError.message}
        onClose={this.props.hideActiveError}
        bullets={[indexedError.occurrences.length + " x occurrences"]}
        actionLabel={"Dismiss"}
        onAction={() => {
          this.props.removeError(activeError.botName, activeError.level, indexedError);
        }}
        onPrevious={onPrevious}
        onNext={onNext}
      />);
    }

    return (
      <div>
        <div className={grid.container}>
          {modalWindow}
          {inner}
        </div>
      </div>
    );
  }
}

export default Bots;
