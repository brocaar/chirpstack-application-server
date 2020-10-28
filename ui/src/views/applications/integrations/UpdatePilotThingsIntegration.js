import React, { Component } from "react";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";

import ApplicationStore from "../../../stores/ApplicationStore";
import PilotThingsIntegrationForm from "./PilotThingsIntegrationForm";


class UpdatePilotThingsIntegration extends Component {
  constructor() {
    super();

    this.state = {};
  }

  onSubmit = (integration) => {
    let integr = integration;
    integr.applicationID = this.props.match.params.applicationID;

    ApplicationStore.updatePilotThingsIntegration(integr, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
    });
  } 

  componentDidMount() {
    ApplicationStore.getPilotThingsIntegration(this.props.match.params.applicationID, (resp) => {
      this.setState({
        object: resp.integration,
      });
    });
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardHeader title="Update Pilot Things integration" />
            <CardContent>
              <PilotThingsIntegrationForm submitLabel="Update integration" onSubmit={this.onSubmit} object={this.state.object} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default UpdatePilotThingsIntegration;
