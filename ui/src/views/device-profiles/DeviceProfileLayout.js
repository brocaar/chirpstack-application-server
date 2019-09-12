import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Grid from '@material-ui/core/Grid';

import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";
import DeviceProfileStore from "../../stores/DeviceProfileStore";
import SessionStore from "../../stores/SessionStore";
import UpdateDeviceProfile from "./UpdateDeviceProfile";


class DeviceProfileLayout extends Component {
  constructor() {
    super();
    this.state = {
      admin: false,
    };
    this.deleteDeviceProfile = this.deleteDeviceProfile.bind(this);
    this.setIsAdmin = this.setIsAdmin.bind(this);
  }

  componentDidMount() {
    DeviceProfileStore.get(this.props.match.params.deviceProfileID, resp => {
      this.setState({
        deviceProfile: resp,
      });
    });

    SessionStore.on("change", this.setIsAdmin);
    this.setIsAdmin();
  }

  componentWillUpdate() {
    SessionStore.removeListener("change", this.setIsAdmin);
  }

  setIsAdmin() {
    this.setState({
      admin: SessionStore.isAdmin() || SessionStore.isOrganizationDeviceAdmin(this.props.match.params.organizationID),
    });
  }

  deleteDeviceProfile() {
    if (window.confirm("Are you sure you want to delete this device-profile?")) {
      DeviceProfileStore.delete(this.props.match.params.deviceProfileID, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/device-profiles`);
      });
    }
  }

  render() {
    if (this.state.deviceProfile === undefined) {
      return(<div></div>);
    }

    let buttons = [];
    if (this.state.admin) {
      buttons = [
          <TitleBarButton
            label="Delete"
            icon={<Delete />}
            color="secondary"
            onClick={this.deleteDeviceProfile}
          />,
      ];
    }

    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={buttons}
        >
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/device-profiles`} title="Device-profiles" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.deviceProfile.deviceProfile.name} />
        </TitleBar>

        <Grid item xs={12}>
          <UpdateDeviceProfile deviceProfile={this.state.deviceProfile.deviceProfile} admin={this.state.admin} />
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(DeviceProfileLayout);
