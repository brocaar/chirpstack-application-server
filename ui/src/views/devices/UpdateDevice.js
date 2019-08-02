import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import DeviceStore from "../../stores/DeviceStore";
import DeviceForm from "./DeviceForm";


const styles = {
  card: {
    overflow: "visible",
  },
};


class UpdateDevice extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(device) {
    DeviceStore.update(device, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}`);
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <DeviceForm
                submitLabel="Update device"
                object={this.props.device}
                onSubmit={this.onSubmit}
                match={this.props.match}
                update={true}
                disabled={!this.props.admin}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(UpdateDevice));
