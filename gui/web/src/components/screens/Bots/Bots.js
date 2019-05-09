import React, { Component } from 'react';
import BotCard from '../../molecules/BotCard/BotCard';
import Button from '../../atoms/Button/Button';
import EmptyList from '../../molecules/EmptyList/EmptyList';
import ScreenHeader from '../../molecules/ScreenHeader/ScreenHeader';
import grid from '../../_styles/grid.module.scss';

import autogenerate from '../../../kelp-ops-api/autogenerate';

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
 
    this.gotoForm = this.gotoForm.bind(this);
    this.gotoDetails = this.gotoDetails.bind(this);
    this.autogenerateBot = this.autogenerateBot.bind(this);
    this.createBot = this.createBot.bind(this);
  }

  componentWillUnmount() {
    if (this._asyncRequest) {
      this._asyncRequest.cancel();
      this._asyncRequest = null;
    }
  }
  
  gotoForm() {
    this.props.history.push('/new')
  }

  gotoDetails() {
    this.props.history.push('/details')
  }

  autogenerateBot() {
    this._asyncRequest = autogenerate(this.props.baseUrl).then(newBot => {
      this._asyncRequest = null;
      this.setState(prevState => ({
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
            icon="download"
            hsize="short"
            onClick={this.autogenerateBot}
          />
          <Button 
            variant="faded" 
            hsize="short"
            icon="add" 
            onClick={this.gotoForm}
            children="New Bots"/>
        </ScreenHeader>
      );

      let cards = this.state.bots.map((bot, index) => (
        <BotCard
          key={index} 
          name={bot.name}
          running={bot.running}
          test={bot.test}
          warnings={bot.warnings}
          errors={bot.errors}
          showDetailsFn={this.gotoDetails}
          baseUrl={this.props.baseUrl}
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
