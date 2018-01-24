import React, { Component } from 'react';
import { Route, Switch } from 'react-router-dom';

import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import Errors from "./components/Errors";
import dispatcher from "./dispatcher";

// users
import Login from "./views/users/Login";
import CreateUser from "./views/users/CreateUser";
import UpdatePassword from "./views/users/UpdatePassword";
import ListUsers from "./views/users/ListUsers";
import UpdateUser from "./views/users/UpdateUser";

// organizations
import OrganizationRedirect from './views/organizations/OrganizationRedirect';
import ListOrganizations from './views/organizations/ListOrganizations';
import CreateOrganization from './views/organizations/CreateOrganization';
import OrganizationLayout from './views/organizations/OrganizationLayout';

// network-servers
import ListNetworkServers from "./views/networkservers/ListNetworkServers";
import CreateNetworkServer from "./views/networkservers/CreateNetworkServer";
import NetworkServerLayout from "./views/networkservers/NetworkServerLayout";
import ChannelConfigurationLayout from "./views/gateways/ChannelConfigurationLayout";

// gateways
import GatewayLayout from "./views/gateways/GatewayLayout";

// applications
import ApplicationLayout from './views/applications/ApplicationLayout';

// devices
import NodeLayout from './views/nodes/NodeLayout';


class Layout extends Component {
  onClick() {
    dispatcher.dispatch({
      type: "BODY_CLICK",
    });
  }

  render() {
    return (
      <div>
        <Navbar />
        <div className="container" onClick={this.onClick}>
          <div className="row">
            <Errors />
            <Switch>
              <Route exact path="/" component={OrganizationRedirect} />
              <Route exact path="/login" component={Login} />
              <Route exact path="/users/create" component={CreateUser} />
              <Route exact path="/users/:userID/password" component={UpdatePassword} />
              <Route exact path="/users/:userID/edit" component={UpdateUser} />
              <Route exact path="/users" component={ListUsers} />

              <Route exact path="/network-servers" component={ListNetworkServers} />
              <Route exact path="/network-servers/create" component={CreateNetworkServer} />
              {/* \d+ regexp to make sure we don't match channel-configurations/create */}
              <Route path="/network-servers/:networkServerID/channel-configurations/:channelConfigurationID(\d+)" component={ChannelConfigurationLayout} />
              <Route path="/network-servers/:networkServerID" component={NetworkServerLayout} />

              <Route exact path="/organizations" component={ListOrganizations} />
              <Route exact path="/organizations/create" component={CreateOrganization} />
              {/* \w{16} to make sure we don't match gateways/create */}
              <Route path="/organizations/:organizationID/gateways/:mac(\w{16})" component={GatewayLayout} />
              {/* \w{16} to make sure we don't match nodes/create */}
              <Route path="/organizations/:organizationID/applications/:applicationID/nodes/:devEUI(\w{16})" component={NodeLayout} />
              {/* \d+ regexp to make sure we don't match 'applications/create' */}
              <Route path="/organizations/:organizationID/applications/:applicationID(\d+)" component={ApplicationLayout} />
              <Route path="/organizations/:organizationID" component={OrganizationLayout} />
            </Switch>
          </div>
        </div>
        <Footer />
      </div>
    );
  }
}

export default Layout;
