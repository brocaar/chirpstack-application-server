import React, { Component } from 'react';
import { Link } from 'react-router';

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
    this.onDelete = this.onDelete.bind(this);
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

  onDelete() {
    if (confirm("Are you sure you want to delete this network-server?")) {
      NetworkServerStore.deleteNetworkServer(this.props.params.networkServerID, (responseData) => {
        this.context.router.push("network-servers");
      });
    }
  }

  render() {
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

export default UpdateNetworkServer;