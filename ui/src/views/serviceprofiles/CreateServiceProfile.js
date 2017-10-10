import React, { Component } from "react";

import ServiceProfileStore from "../../stores/ServiceProfileStore";
import ServiceProfileForm from "../../components/ServiceProfileForm";


class CreateServiceProfile extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

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
      this.context.router.push("organizations/"+this.props.params.organizationID+"/service-profiles");
    });
  }

  componentDidMount() {
    this.setState({
      serviceProfile: {
        organizationID: this.props.params.organizationID,
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
          <ServiceProfileForm serviceProfile={this.state.serviceProfile} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default CreateServiceProfile;