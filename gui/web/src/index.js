import React from 'react';
import ReactDOM from 'react-dom';
import './index.scss';
import App from './App';
import { Auth0Provider } from "@auth0/auth0-react";
import * as serviceWorker from './serviceWorker';
import config from './auth0-config.json'

const auth0enabled = config.auth0_enabled;

ReactDOM.render(
  <div>
    {auth0enabled ? (<Auth0Provider
  domain= {config.domain}
  clientId= {config.client_id}
  redirectUri= {window.location.origin}
  audience= {config.audience}
>
  <App />
  </Auth0Provider>) : (<App />)}
    </div>,
document.getElementById('root')
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
