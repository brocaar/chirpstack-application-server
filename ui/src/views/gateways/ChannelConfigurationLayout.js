import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from 'react-router-dom';

import NetworkServerStore from "../../stores/NetworkServerStore";
import GatewayStore from "../../stores/GatewayStore";

import UpdateChannelConfiguration from "./UpdateChannelConfiguration";
import UpdateChannelConfigurationExtraChannels from "./UpdateChannelConfigurationExtraChannels";


class ChannelConfigurationLayout extends Component {
  constructor() {
    super();

    this.state = {
      networkServer: {},
      configuration: {},
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.getNetworkServer(this.props.match.params.networkServerID, (networkServer) => {
      this.setState({
        networkServer: networkServer,
      });
    });

    GatewayStore.getChannelConfiguration(this.props.match.params.networkServerID, this.props.match.params.channelConfigurationID, (configuration) => {
      this.setState({
        configuration: configuration,
      });
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this channel-configuration?")) {
      GatewayStore.deleteChannelConfiguration(this.props.match.params.networkServerID, this.props.match.params.channelConfigurationID, (responseData) => {
        this.props.history.push(`/network-servers/${this.props.match.params.networkServerID}/channel-configurations`);
      });
    }
  }

  render() {
    let activeTab = this.props.location.pathname.replace(this.props.match.url, '');

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/network-servers">Network-servers</Link></li>
          <li><Link to={`/network-servers/${this.state.networkServer.id}`}>{this.state.networkServer.name}</Link></li>
          <li><Link to={`/network-servers/${this.state.networkServer.id}/channel-configurations`}>Channel-configurations</Link></li>
          <li className="active">{this.state.configuration.name}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <a><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete channel-configuration</button></a>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={activeTab === "/edit" ? "active" : ""}><Link to={`/network-servers/${this.props.match.params.networkServerID}/channel-configurations/${this.props.match.params.channelConfigurationID}/edit`}>Channel-configuration</Link></li>
          <li role="presentation" className={activeTab === "/extra-channels" ? "active" : ""}><Link to={`/network-servers/${this.props.match.params.networkServerID}/channel-configurations/${this.props.match.params.channelConfigurationID}/extra-channels`}>Extra channels</Link></li>
        </ul>
        <hr />
        <Switch>
          <Route path={`${this.props.match.path}/edit`} component={UpdateChannelConfiguration} />
          <Route path={`${this.props.match.path}/extra-channels`} component={UpdateChannelConfigurationExtraChannels} />
        </Switch>
      </div>
    );
  }
}

export default withRouter(ChannelConfigurationLayout);
