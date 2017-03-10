import React from 'react';
import ReactDOM from 'react-dom';
import { Router, Route, IndexRoute, hashHistory } from 'react-router';

import Layout from './Layout';

// applications
import ListApplications from './views/applications/ListApplications';
import CreateApplication from "./views/applications/CreateApplication";
import UpdateApplication from "./views/applications/UpdateApplication";

// nodes
import ListNodes from './views/nodes/ListNodes';
import UpdateNode from './views/nodes/UpdateNode';
import CreateNode from "./views/nodes/CreateNode";
import ActivateNode from "./views/nodes/ActivateNode";

// channels
import ChannelLists from "./views/channels/ChannelLists";
import ChannelListDetails from "./views/channels/ChannelListDetails";
import CreateChannelList from "./views/channels/CreateChannelList";

// users
import Login from "./views/users/Login";

// styling
import 'bootstrap/dist/css/bootstrap.css';
import 'bootswatch/paper/bootstrap.css';
import './index.css';


ReactDOM.render(
  <Router history={hashHistory}>
    <Route path="/" component={Layout}>
      <IndexRoute component={ListApplications}></IndexRoute>
      <Route path="applications" component={ListApplications}></Route>
      <Route path="applications/create" component={CreateApplication}></Route>
      <Route path="applications/:applicationID/edit" component={UpdateApplication}></Route>
      <Route path="applications/:applicationID/nodes/create" component={CreateNode}></Route>
      <Route path="applications/:applicationID/nodes/:devEUI/edit" component={UpdateNode}></Route>
      <Route path="applications/:applicationID/nodes/:devEUI/activation" component={ActivateNode}></Route>
      <Route path="applications/:applicationID" component={ListNodes}></Route>
      <Route path="channels" component={ChannelLists}></Route>
      <Route path="channels/create" component={CreateChannelList}></Route>
      <Route path="channels/:id" component={ChannelListDetails}></Route>
      <Route path="login" component={Login}></Route>
    </Route>
  </Router>,
  document.getElementById('root')
);
