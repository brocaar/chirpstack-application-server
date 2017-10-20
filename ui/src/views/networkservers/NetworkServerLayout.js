import React, { Component } from "react";
import { Link } from 'react-router';

import NetworkServerStore from "../../stores/NetworkServerStore";


class NetworkServerLayout extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      networkServer: {},
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.getNetworkServer(this.props.params.networkServerID, (networkServer) => {
      this.setState({
        networkServer: networkServer,
      });
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this network-server?")) {
      NetworkServerStore.deleteNetworkServer(this.props.params.networkServerID, (responseData) => {
        this.context.router.push("network-servers");
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
          <li><Link to="network-servers">Network servers</Link></li>
          <li className="active">{this.state.networkServer.name}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete network-server</button></Link>
          </div>
        </div>
        <div className="nav nav-tabs">
          <li role="presentation" className={(activeTab === "" ? "active" : "")}><Link to={`network-servers/${this.props.params.networkServerID}`}>Network-server configuration</Link></li>
          <li role="presentation" className={(activeTab.startsWith("channel-configurations") ? "active": "")}><Link to={`network-servers/${this.props.params.networkServerID}/channel-configurations`}>Channel configurations</Link></li>
        </div>
        <hr />
        {this.props.children}
      </div>
    );
  }
}

export default NetworkServerLayout;