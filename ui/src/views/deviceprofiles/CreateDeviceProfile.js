import React, { Component } from "react";

import DeviceProfileStore from "../../stores/DeviceProfileStore";
import DeviceProfileForm from "../../components/DeviceProfileForm";


class CreateDeviceProfile extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

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
      this.context.router.push("organizations/"+this.props.params.organizationID+"/device-profiles");
    });
  }

  componentDidMount() {
    this.setState({
      deviceProfile: {
        organizationID: this.props.params.organizationID,
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
          <DeviceProfileForm organizationID={this.props.params.organizationID} deviceProfile={this.state.deviceProfile} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default CreateDeviceProfile;