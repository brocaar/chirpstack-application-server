import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import GatewayProfileStore from "../../stores/GatewayProfileStore";
import GatewayProfileForm from "../../components/GatewayProfileForm";


class CreateGatewayProfile extends Component {
  constructor() {
    super();

    this.state = {
      gatewayProfile: {
        gatewayProfile: {},
      },
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    this.setState({
      gatewayProfile: {
        networkServerID: this.props.match.params.networkServerID,
        gatewayProfile: {},
      },
    });
  }

  onSubmit(gatewayProfile) {
    GatewayProfileStore.createGatewayProfile(gatewayProfile, (responseData) => {
      this.props.history.push(`/network-servers/${this.props.match.params.networkServerID}/gateway-profiles`);
    });
  }

  render() {
    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-heading">
            Create gateway-profile
          </div>
          <div className="panel-body">
            <GatewayProfileForm profile={this.state.gatewayProfile} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(CreateGatewayProfile);
