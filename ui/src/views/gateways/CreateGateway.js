import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import GatewayStore from "../../stores/GatewayStore";
import GatewayForm from "../../components/GatewayForm";

class CreateGateway extends Component {
  constructor() {
    super();

    this.state = {
      gateway: {
        location: {},
      },
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    this.setState({
      gateway: {location: {}, organizationID: this.props.match.params.organizationID},
    });
  }

  onSubmit(gateway) {
    GatewayStore.createGateway({gateway: gateway}, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/gateways`);
    });
  }

  render() {
    return(
        <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title">Create gateway</h3>
        </div>
        <div className="panel-body">
          <GatewayForm organizationID={this.props.match.params.organizationID} gateway={this.state.gateway} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(CreateGateway);
