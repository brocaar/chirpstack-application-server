import React, { Component } from 'react';
import { Link } from 'react-router';

import GatewayStore from "../../stores/GatewayStore";
import GatewayForm from "../../components/GatewayForm";

class UpdateGateway extends Component {
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

  componentWillMount() {
    GatewayStore.getGateway(this.props.params.mac, (gateway) => {
      this.setState({
        gateway: gateway,
      });
    });
  }

  onSubmit(gateway) {
    GatewayStore.updateGateway(this.props.params.mac, gateway, (responseData) => {
      this.context.router.push('gateways/'+this.props.params.mac);
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/gateways">Gateways</Link></li>
          <li><Link to={`/gateways/${this.props.params.mac}`}>{this.state.gateway.name}</Link></li>
          <li className="active">Edit gateway</li>
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

export default UpdateGateway;
