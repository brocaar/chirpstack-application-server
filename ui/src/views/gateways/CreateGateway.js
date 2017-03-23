import React, { Component } from "react";
import { Link } from "react-router";

import GatewayStore from "../../stores/GatewayStore";
import GatewayForm from "../../components/GatewayForm";

class CreateGateway extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      gateway: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(gateway) {
    GatewayStore.createGateway(gateway, (responseData) => {
      this.context.router.push("/gateways");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="">Dashboard</Link></li>
          <li><Link to="gateways">Gateways</Link></li>
          <li className="active">Create gateway</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <GatewayForm gateway={this.state.gateway} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateGateway;
