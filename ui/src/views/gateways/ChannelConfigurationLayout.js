import React, { Component } from "react";
import { Link } from 'react-router';

import NetworkServerStore from "../../stores/NetworkServerStore";
import GatewayStore from "../../stores/GatewayStore";


class ChannelConfigurationLayout extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      networkServer: {},
      configuration: {},
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.getNetworkServer(this.props.params.networkServerID, (networkServer) => {
      this.setState({
        networkServer: networkServer,
      });
    });

    GatewayStore.getChannelConfiguration(this.props.params.networkServerID, this.props.params.channelConfigurationID, (configuration) => {
      this.setState({
        configuration: configuration,
      });
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this channel-configuration?")) {
      GatewayStore.deleteChannelConfiguration(this.props.params.networkServerID, this.props.params.channelConfigurationID, (responseData) => {
        this.context.router.push(`network-servers/${this.props.params.networkServerID}/channel-configurations`);
      });
    }
  }

  render() {
    let activeTab = "";

    if (typeof(this.props.children.props.route.path) !== "undefined") {
      activeTab = this.props.children.props.route.path; 
    }

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="network-servers">Network-servers</Link></li>
          <li><Link to={`network-servers/${this.state.networkServer.id}`}>{this.state.networkServer.name}</Link></li>
          <li><Link to={`network-servers/${this.state.networkServer.id}/channel-configurations`}>Channel-configurations</Link></li>
          <li className="active">{this.state.configuration.name}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete channel-configuration</button></Link>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={activeTab === "edit" ? "active" : ""}><Link to={`network-servers/${this.props.params.networkServerID}/channel-configurations/${this.props.params.channelConfigurationID}/edit`}>Channel-configuration</Link></li>
          <li role="presentation" className={activeTab === "extra-channels" ? "active" : ""}><Link to={`network-servers/${this.props.params.networkServerID}/channel-configurations/${this.props.params.channelConfigurationID}/extra-channels`}>Extra channels</Link></li>
        </ul>
        <hr />
        {this.props.children}
      </div>
    );
  }
}

export default ChannelConfigurationLayout;
