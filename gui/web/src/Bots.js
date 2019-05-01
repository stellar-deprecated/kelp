import React, { Component } from 'react';
import BotCard from './components/molecules/BotCard/BotCard';
import Button from './components/atoms/Button/Button';
import EmptyList from './components/molecules/EmptyList/EmptyList';
import ScreenHeader from './components/molecules/ScreenHeader/ScreenHeader';
import { BrowserRouter as Router, Route, Link } from "react-router-dom";

import grid from './components/_settings/grid.module.scss';
import Modal from './components/molecules/Modal/Modal';
import Welcome from './components/molecules/Welcome/Welcome';


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
    this.state = { bots: [] };
 
    this.createBot = this.createBot.bind(this);
    this.gotoForm = this.gotoForm.bind(this);
  }
  
  gotoForm() {
    this.props.history.push('/form')
  }
  
  createBot () {
    let rand = Math.floor(Math.random() * placeaholderBots.length);
    let newElement = placeaholderBots[rand];
    this.setState(prevState => ({
      bots: [...prevState.bots, newElement]
    }))
  }

  render() {

    return (
      <div>
        <div className={grid.container}> 
          { this.state.bots.length ? (
            <div>
              <ScreenHeader title={'My Bots'}>
                <Button 
                  variant="faded"
                  icon="download"
                  hsize="short"
                  onClick={this.createBot}
                > 
                </Button>
                <Button 
                  variant="faded" 
                  hsize="short"
                  icon="add" 
                  onClick={this.gotoForm}>New Bot
                </Button>
              </ScreenHeader>
              {this.state.bots.map(bot => (
                <BotCard 
                  name={bot.name}
                  running={bot.running}
                  test={bot.test}
                  warnings={bot.warnings}
                  errors={bot.errors}
                />
              ))}
          </div>
          ) : (
            <EmptyList onClick={this.createBot}/>
          )}
        </div>

        <Modal 
          type="error"
          title="Harry the Green Plankton has two warnings:"
          actionLabel="Go to bot settings"
          bullets={['Funds are low', 'Another warning example']}
        />

        <Welcome/>
      </div>
    );
  }
}

export default Bots;
