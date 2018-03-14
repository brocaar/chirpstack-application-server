import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from 'react-router-dom';

import NetworkServerStore from "../../stores/NetworkServerStore";

// network-servers
import UpdateNetworkServer from "./UpdateNetworkServer";

// gateways
import ListChannelConfigurations from "../gateways/ListChannelConfigurations";
import CreateChannelConfiguration from "../gateways/CreateChannelConfiguration";


class NetworkServerLayout extends Component {
  constructor() {
    super();

    this.state = {
      networkServer: {},
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.getNetworkServer(this.props.match.params.networkServerID, (networkServer) => {
      this.setState({
        networkServer: networkServer,
      });
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this network-server?")) {
      NetworkServerStore.deleteNetworkServer(this.props.match.params.networkServerID, (responseData) => {
        this.props.history.push("network-servers");
      });
    }
  }

  render() {
    let activeTab = this.props.location.pathname.replace(this.props.match.url, '');

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/network-servers">Network servers</Link></li>
          <li className="active">{this.state.networkServer.name} ({this.state.networkServer.region} @ {this.state.networkServer.version})</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <a><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete network-server</button></a>
          </div>
        </div>
        <div className="nav nav-tabs">
          <li role="presentation" className={(activeTab === "" ? "active" : "")}><Link to={`/network-servers/${this.props.match.params.networkServerID}`}>Network-server configuration</Link></li>
          <li role="presentation" className={(activeTab.startsWith("/channel-configurations") ? "active": "")}><Link to={`/network-servers/${this.props.match.params.networkServerID}/channel-configurations`}>Channel configurations</Link></li>
        </div>
        <hr />

        <Switch>
          <Route exact path={this.props.match.path} component={UpdateNetworkServer} />
          <Route path={`${this.props.match.path}/channel-configurations/create`} component={CreateChannelConfiguration} />
          <Route path={`${this.props.match.path}/channel-configurations`} component={ListChannelConfigurations} />
        </Switch>
      </div>
    );
  }
}

export default withRouter(NetworkServerLayout);