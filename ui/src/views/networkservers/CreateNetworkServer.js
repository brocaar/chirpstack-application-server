import React, { Component } from "react";
import { Link, withRouter } from "react-router-dom";

import NetworkServerStore from "../../stores/NetworkServerStore";
import NetworkServerForm from "../../components/NetworkServerForm";


class CreateNetworkServer extends Component {
  constructor() {
    super();

    this.state = {
      networkServer: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(networkServer) {
    NetworkServerStore.createNetworkServer(networkServer, (responseData) => {
      this.props.history.push("/network-servers");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/network-servers">Network servers</Link></li>
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

export default withRouter(CreateNetworkServer);
