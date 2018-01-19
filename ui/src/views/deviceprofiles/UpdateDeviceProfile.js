import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import DeviceProfileStore from "../../stores/DeviceProfileStore";
import SessionStore from "../../stores/SessionStore";
import DeviceProfileForm from "../../components/DeviceProfileForm";


class UpdateDeviceProfile extends Component {
  constructor() {
    super();

    this.state = {
      deviceProfile: {
          deviceProfile: {},
      },
      isAdmin: false,
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    DeviceProfileStore.getDeviceProfile(this.props.match.params.deviceProfileID, (deviceProfile) => {
      this.setState({
        deviceProfile: deviceProfile,
        isAdmin: SessionStore.isAdmin(),
      });
    });
  }

  onSubmit(deviceProfile) {
    DeviceProfileStore.updateDeviceProfile(deviceProfile.deviceProfile.deviceProfileID, deviceProfile, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/device-profiles`);
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this device-profile?")) {
      DeviceProfileStore.deleteDeviceProfile(this.props.match.params.deviceProfileID, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/device-profiles`);
      });
    }
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <h3 className="panel-title panel-title-buttons pull-left">Update device-profile</h3>
          <div className={"btn-group pull-right " + (this.state.isAdmin ? "" : "hidden")}>
            <a><button type="button" className="btn btn-danger btn-sm" onClick={this.onDelete}>Remove device-profile</button></a>
          </div>
        </div>
        <div className="panel-body">
          <DeviceProfileForm organizationID={this.props.match.params.organizationID} deviceProfile={this.state.deviceProfile} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateDeviceProfile);