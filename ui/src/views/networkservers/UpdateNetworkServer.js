import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import NetworkServerStore from "../../stores/NetworkServerStore";
import NetworkServerForm from "../../components/NetworkServerForm";


class UpdateNetworkServer extends Component {
  constructor() {
    super();

    this.state = {
      networkServer: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.getNetworkServer(this.props.match.params.networkServerID, (networkServer) => {
      this.setState({
        networkServer: networkServer,
      });
    });
  }

  onSubmit(networkServer) {
    NetworkServerStore.updateNetworkServer(this.props.match.params.networkServerID, networkServer, (responseData) => {
      this.props.history.push("/network-servers");
    });
  }

  render() {
    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-body">
            <NetworkServerForm networkServer={this.state.networkServer} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateNetworkServer);