import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";

import ApplicationStore from "../../stores/ApplicationStore";
import DeviceStore from "../../stores/DeviceStore";
import FUOTADeploymentStore from "../../stores/FUOTADeploymentStore";
import FUOTADeploymentForm from "./FUOTADeploymentForm";


const styles = {
  card: {
    overflow: "visible",
  },
};


class CreateFUOTADeploymentForDevice extends Component {
  constructor() {
    super();
    this.state = {};
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });

    DeviceStore.get(this.props.match.params.devEUI, resp => {
      this.setState({
        device: resp,
      });
    });
  }

  onSubmit(fuotaDeployment) {
    FUOTADeploymentStore.createForDevice(this.props.match.params.devEUI, fuotaDeployment, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/fuota-deployments`);
    });
  }

  render() {
    if (this.state.application === undefined || this.state.device === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Applications" to={`/organizations/${this.props.match.params.organizationID}/applications`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.application.application.name} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Devices" to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.device.device.name} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Firmware" to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/fuota-deployments`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Create update job" />
        </TitleBar>

        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <FUOTADeploymentForm
                submitLabel="Create FUOTA deployment"
                onSubmit={this.onSubmit}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(CreateFUOTADeploymentForDevice));

