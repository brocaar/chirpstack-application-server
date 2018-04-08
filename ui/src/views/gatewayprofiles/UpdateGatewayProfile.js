import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import GatewayProfileStore from "../../stores/GatewayProfileStore";
import GatewayProfileForm from "../../components/GatewayProfileForm";


class UpdateGatewayProfile extends Component {
  constructor() {
    super();

    this.state = {
      gatewayProfile: {
        gatewayProfile: {},
      },
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    GatewayProfileStore.getGatewayProfile(this.props.match.params.gatewayProfileID, (gatewayProfile) => {
      this.setState({
        gatewayProfile: gatewayProfile,
      });
    });
  }

  onSubmit(gatewayProfile) {
    GatewayProfileStore.updateGatewayProfile(this.props.match.params.gatewayProfileID, gatewayProfile, (responseData) => {
      this.props.history.push(`/network-servers/${this.props.match.params.networkServerID}/gateway-profiles`);
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this gateway-profile?")) {
      GatewayProfileStore.deleteGatewayProfile(this.props.match.params.gatewayProfileID, (responseData) => {
        this.props.history.push(`/network-servers/${this.props.match.params.networkServerID}/gateway-profiles`);
      });
    }
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <h3 className="panel-title panel-title-buttons pull-left">Update gateway-profile</h3>
          <div className="btn-group pull-right">
            <a><button type="button" className="btn btn-danger btn-sm" onClick={this.onDelete}>Remove gateway-profile</button></a>
          </div>
        </div>
        <div className="panel-body">
          <GatewayProfileForm profile={this.state.gatewayProfile} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateGatewayProfile);
