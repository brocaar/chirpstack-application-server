import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import ServiceProfileStore from "../../stores/ServiceProfileStore";
import ServiceProfileForm from "../../components/ServiceProfileForm";


class CreateServiceProfile extends Component {
  constructor() {
    super();

    this.state = {
      serviceProfile: {
        serviceProfile: {},
      },
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(serviceProfile) {
    ServiceProfileStore.createServiceProfile(serviceProfile, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/service-profiles`);
    });
  }

  componentDidMount() {
    this.setState({
      serviceProfile: {
        organizationID: this.props.match.params.organizationID,
        serviceProfile: {},
      },
    });
  }

  render() {
    return (
      <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title panel-title-buttons">Create service-profile</h3>
        </div>
        <div className="panel-body">
          <ServiceProfileForm organizationID={this.props.match.params.organizationID} serviceProfile={this.state.serviceProfile} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(CreateServiceProfile);
