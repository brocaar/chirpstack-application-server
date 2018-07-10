import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import DeviceProfileStore from "../../stores/DeviceProfileStore";
import DeviceProfileForm from "./DeviceProfileForm";


class UpdateDeviceProfile extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(deviceProfile) {
    DeviceProfileStore.update(deviceProfile, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/device-profiles`);
    });
  }

  render() {
    return(
      <Grid container spacing={24}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <DeviceProfileForm
                submitLabel="Update device-profile"
                object={this.props.deviceProfile}
                onSubmit={this.onSubmit}
                match={this.props.match}
                disabled={!this.props.admin}
                update={true}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(UpdateDeviceProfile);
