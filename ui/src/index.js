import React from 'react';
import ReactDOM from 'react-dom';
import { Router, Route, IndexRoute, hashHistory } from 'react-router';

import Layout from './Layout';
import ListNodes from './views/nodes/ListNodes';
import NodeDetails from './views/nodes/NodeDetails';
import CreateNode from "./views/nodes/CreateNode";

import 'bootstrap/dist/css/bootstrap.css';
import 'bootswatch/paper/bootstrap.css';
import './index.css';


ReactDOM.render(
  <Router history={hashHistory}>
    <Route path="/" component={Layout}>
      <IndexRoute component={ListNodes}></IndexRoute>
      <Route path="nodes/create" component={CreateNode}></Route>
      <Route path="nodes/:devEUI" component={NodeDetails}></Route>
    </Route>
  </Router>,
  document.getElementById('root')
);
