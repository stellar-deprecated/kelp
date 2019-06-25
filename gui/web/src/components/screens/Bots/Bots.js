import React, { Component } from 'react';
import BotCard from '../../molecules/BotCard/BotCard';
import Button from '../../atoms/Button/Button';
import EmptyList from '../../molecules/EmptyList/EmptyList';
import ScreenHeader from '../../molecules/ScreenHeader/ScreenHeader';
import grid from '../../_styles/grid.module.scss';

import autogenerate from '../../../kelp-ops-api/autogenerate';
import listBots from '../../../kelp-ops-api/listBots';

const placeaholderBots = [
  {
    name: 'Harry the Green Plankton',
    running: true,
    test: false,
    warnings: 2,
    errors: 1,
  },
  {
    name: 'Sally the Blue Eel',
    running: false,
    test: true,
    warnings: 2,
    errors: 1,
  },
  {
    name: 'Bruno the Yellow Seaweed',
    running: false,
    test: true,
    warnings: 0,
    errors: 0,
  }
]

class Bots extends Component {
  constructor(props) {
    super(props);
    this.state = {
      bots: [],
    };
 
    this.fetchBots = this.fetchBots.bind(this);
    this.gotoForm = this.gotoForm.bind(this);
    this.gotoDetails = this.gotoDetails.bind(this);
    this.autogenerateBot = this.autogenerateBot.bind(this);
    this.createBot = this.createBot.bind(this);
    
    this._asyncRequests = {};
  }

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
  
  gotoForm() {
    this.props.history.push('/new')
  }

  gotoDetails() {
    this.props.history.push('/details')
  }

  fetchBots() {
    var _this = this
    this._asyncRequests["listBots"] = listBots(this.props.baseUrl).then(bots => {
      if (!_this._asyncRequests["listBots"]) {
        // if it has been deleted it means we don't want to process the result
        return
      }

      delete _this._asyncRequests["listBots"];
      if (bots.hasOwnProperty('error')) {
        console.log("error in listBots: " + bots.error);
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
    let rand = Math.floor(Math.random() * placeaholderBots.length);
    let newElement = placeaholderBots[rand];
    this.setState(prevState => ({
      bots: [...prevState.bots, newElement]
    }))
  }

  render() {
    let inner = <EmptyList autogenerateFn={this.autogenerateBot} createBotFn={this.createBot}/>;
    if (this.state.bots.length) {
      let screenHeader = (
        <ScreenHeader title={'My Bots'}>
          <Button 
            variant="faded" 
            hsize="short"
            icon="add" 
            onClick={this.gotoForm}
            >
              New Bot
            </Button>
        </ScreenHeader>
      );

      let cards = this.state.bots.map((bot, index) => (
        <BotCard
          key={index} 
          name={bot.name}
          history={this.props.history}
          running={bot.running}
          test={bot.test}
          warnings={bot.warnings}
          errors={bot.errors}
          showDetailsFn={this.gotoDetails}
          baseUrl={this.props.baseUrl}
          reload={this.fetchBots}
        />
      ));

      inner = (
        <div>
          {screenHeader}
          {cards}
        </div>
      );
    }

    return (
      <div>
        <div className={grid.container}> 
          {inner}
        </div>
      </div>
    );
  }
}

export default Bots;
