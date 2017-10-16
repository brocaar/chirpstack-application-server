import React, { Component } from "react";
import { Link } from 'react-router';

import ServiceProfileStore from "../../stores/ServiceProfileStore";
import SessionStore from "../../stores/SessionStore";
import ServiceProfileForm from "../../components/ServiceProfileForm";


class UpdateServiceProfile extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      serviceProfile: {
          serviceProfile: {},
      },
      isAdmin: false,
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    ServiceProfileStore.getServiceProfile(this.props.params.serviceProfileID, (serviceProfile) => {
      this.setState({
        serviceProfile: serviceProfile,
        isAdmin: SessionStore.isAdmin(),
      });
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: SessionStore.isAdmin(),
      });
    });
  }

  onSubmit(serviceProfile) {
    ServiceProfileStore.updateServiceProfile(serviceProfile.serviceProfile.serviceProfileID, serviceProfile, (responseData) => {
      this.context.router.push("organizations/"+this.props.params.organizationID+"/service-profiles")
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this service-profile?")) {
      ServiceProfileStore.deleteServiceProfile(this.props.params.serviceProfileID, (responseData) => {
        this.context.router.push("organizations/"+this.props.params.organizationID+"/service-profiles");
      });
    }
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <h3 className="panel-title panel-title-buttons pull-left">Update service-profile</h3>
          <div className={"btn-group pull-right " + (this.state.isAdmin ? "" : "hidden")}>
            <Link><button type="button" className="btn btn-danger btn-sm" onClick={this.onDelete}>Remove service-profile</button></Link>
          </div>
        </div>
        <div className="panel-body">
          <ServiceProfileForm organizationID={this.props.params.organizationID} serviceProfile={this.state.serviceProfile} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default UpdateServiceProfile;