import React, { Component } from "react";
import { Link } from "react-router";

import NetworkServerStore from "../../stores/NetworkServerStore";
import NetworkServerForm from "../../components/NetworkServerForm";


class CreateNetworkServer extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      networkServer: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(networkServer) {
    NetworkServerStore.createNetworkServer(networkServer, (responseData) => {
      this.context.router.push("network-servers");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="network-servers">Network servers</Link></li>
          <li className="active">Add network-server</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NetworkServerForm networkServer={this.state.networkServer} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateNetworkServer;