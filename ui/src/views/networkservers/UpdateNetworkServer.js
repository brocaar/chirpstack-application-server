import React, { Component } from 'react';

import NetworkServerStore from "../../stores/NetworkServerStore";
import NetworkServerForm from "../../components/NetworkServerForm";


class UpdateNetworkServer extends Component {
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

  componentDidMount() {
    NetworkServerStore.getNetworkServer(this.props.params.networkServerID, (networkServer) => {
      this.setState({
        networkServer: networkServer,
      });
    });
  }

  onSubmit(networkServer) {
    NetworkServerStore.updateNetworkServer(this.props.params.networkServerID, networkServer, (responseData) => {
      this.context.router.push("network-servers");
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

export default UpdateNetworkServer;