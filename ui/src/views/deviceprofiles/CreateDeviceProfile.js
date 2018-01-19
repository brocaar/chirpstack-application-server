import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import DeviceProfileStore from "../../stores/DeviceProfileStore";
import DeviceProfileForm from "../../components/DeviceProfileForm";


class CreateDeviceProfile extends Component {
  constructor() {
    super();

    this.state = {
      deviceProfile: {
        deviceProfile: {},
      },
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(deviceProfile) {
    DeviceProfileStore.createDeviceProfile(deviceProfile, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/device-profiles`);
    });
  }

  componentDidMount() {
    this.setState({
      deviceProfile: {
        organizationID: this.props.match.params.organizationID,
        deviceProfile: {},
      },
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title panel-title-buttons">Create device-profile</h3>
        </div>
        <div className="panel-body">
          <DeviceProfileForm organizationID={this.props.match.params.organizationID} deviceProfile={this.state.deviceProfile} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(CreateDeviceProfile);
