import React from 'react';
import ReactDOM from 'react-dom';
import { Router, Route, IndexRoute, hashHistory } from 'react-router';

import Layout from './Layout';

// stores
import ErrorStore from "./stores/ErrorStore";

// applications
import ListApplications from './views/applications/ListApplications';
import CreateApplication from "./views/applications/CreateApplication";
import UpdateApplication from "./views/applications/UpdateApplication";
import ApplicationUsers from "./views/applications/ApplicationUsers";
import CreateApplicationUser from "./views/applications/CreateApplicationUser";
import UpdateApplicationUser from "./views/applications/UpdateApplicationUser";

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
import CreateUser from "./views/users/CreateUser";
import ListUsers from "./views/users/ListUsers";
import UpdateUser from "./views/users/UpdateUser";
import UpdatePassword from "./views/users/UpdatePassword";

// gateways
import ListGateways from "./views/gateways/ListGateways";
import GatewayDetails from "./views/gateways/GatewayDetails";
import CreateGateway from "./views/gateways/CreateGateway";
import UpdateGateway from "./views/gateways/UpdateGateway";

// organizations
import ListOrganizations from './views/organizations/ListOrganizations';
import CreateOrganization from './views/organizations/CreateOrganization';
import UpdateOrganization from './views/organizations/UpdateOrganization';
import OrganizationRedirect from './views/organizations/OrganizationRedirect';
import OrganizationUsers from './views/organizations/OrganizationUsers';
import CreateOrganizationUser from './views/organizations/CreateOrganizationUser';
import UpdateOrganizationUser from './views/organizations/UpdateOrganizationUser';

// fix leaflet image source
import Leaflet from 'leaflet';
Leaflet.Icon.Default.imagePath = '//cdnjs.cloudflare.com/ajax/libs/leaflet/1.0.0/images/'

// styling
import 'bootstrap/dist/css/bootstrap.css';
import 'bootswatch/paper/bootstrap.css';
import 'react-select/dist/react-select.css';
import 'leaflet/dist/leaflet.css';
import './index.css';


ReactDOM.render(
  <Router history={hashHistory}>
    <Route path="/" component={Layout} onChange={clearErrors}>
      <IndexRoute component={OrganizationRedirect}></IndexRoute>
      <Route path="login" component={Login}></Route>
      <Route path="users/create" component={CreateUser}></Route>
      <Route path="users/:userID/edit" component={UpdateUser}></Route>
      <Route path="users/:userID/password" component={UpdatePassword}></Route>
      <Route path="users" component={ListUsers}></Route>
      <Route path="channels" component={ChannelLists}></Route>
      <Route path="channels/create" component={CreateChannelList}></Route>
      <Route path="channels/:id" component={ChannelListDetails}></Route>
      <Route path="organizations" component={ListOrganizations}></Route>
      <Route path="organizations/create" component={CreateOrganization}></Route>
      <Route path="organizations/:id/edit" component={UpdateOrganization}></Route>
      <Route path="organizations/:organizationID" component={ListApplications}></Route>
      <Route path="organizations/:organizationID/applications" component={ListApplications}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/edit" component={UpdateApplication}></Route>
      <Route path="organizations/:organizationID/applications/create" component={CreateApplication}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/edit" component={UpdateApplication}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID" component={ListNodes}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/users" component={ApplicationUsers}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/users/create" component={CreateApplicationUser}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/users/:userID/edit" component={UpdateApplicationUser}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/nodes/create" component={CreateNode}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/nodes/:devEUI/edit" component={UpdateNode}></Route>
      <Route path="organizations/:organizationID/applications/:applicationID/nodes/:devEUI/activation" component={ActivateNode}></Route>
      <Route path="organizations/:organizationID/gateways" component={ListGateways}></Route>
      <Route path="organizations/:organizationID/gateways/create" component={CreateGateway}></Route>
      <Route path="organizations/:organizationID/gateways/:mac" component={GatewayDetails}></Route>
      <Route path="organizations/:organizationID/gateways/:mac/edit" component={UpdateGateway}></Route>
      <Route path="organizations/:organizationID/users" component={OrganizationUsers}></Route>
      <Route path="organizations/:organizationID/users/create" component={CreateOrganizationUser}></Route>
      <Route path="organizations/:organizationID/users/:userID/edit" component={UpdateOrganizationUser}></Route>
    </Route>
  </Router>,
  document.getElementById('root')
);

function clearErrors(prevRoute, nextRoute) {
  ErrorStore.clear();  
}
