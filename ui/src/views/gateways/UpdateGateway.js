import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import GatewayStore from "../../stores/GatewayStore";
import GatewayForm from "../../components/GatewayForm";


class UpdateGateway extends Component {
  constructor() {
    super();

    this.state = {
      gateway: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    GatewayStore.getGateway(this.props.match.params.mac, (gateway) => {
      this.setState({
        gateway: gateway,
      });
    });
  }

  onSubmit(gateway) {
    GatewayStore.updateGateway(this.props.match.params.mac, gateway, (responseData) => {
      this.props.history.push(`/organizations/${gateway.organizationID}/gateways/${gateway.mac}`);
      window.scrollTo(0, 0);
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <GatewayForm organizationID={this.props.match.params.organizationID} gateway={this.state.gateway} onSubmit={this.onSubmit} update={true} />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateGateway);
