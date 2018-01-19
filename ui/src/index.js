// styling
import 'bootstrap/dist/css/bootstrap.css';
import 'bootswatch/paper/bootstrap.css';
import 'react-select/dist/react-select.css';
import 'leaflet/dist/leaflet.css';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/base16-light.css';
import './index.css';

import React from 'react';
import ReactDOM from 'react-dom';
import { Router, Route } from 'react-router-dom';
// import { BrowserRouter, Route, IndexRoute } from 'react-router-dom';

import Layout from './Layout';
import history from './history';

// stores
import ErrorStore from "./stores/ErrorStore";

// fix leaflet image source
import Leaflet from 'leaflet';
Leaflet.Icon.Default.imagePath = '//cdnjs.cloudflare.com/ajax/libs/leaflet/1.0.0/images/'



ReactDOM.render(
  <Router history={history}>
    <Route path="" component={Layout} onChange={clearErrors} />
  </Router>,
  document.getElementById('root')
);

function clearErrors(prevRoute, nextRoute) {
  ErrorStore.clear();  
}
