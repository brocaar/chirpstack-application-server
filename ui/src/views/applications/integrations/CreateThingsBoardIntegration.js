import React, { Component } from "react";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";

import ApplicationStore from "../../../stores/ApplicationStore";
import ThingsBoardIntegrationForm from "./ThingsBoardIntegrationForm";


class CreateThingsBoardIntegration extends Component {
  onSubmit = (integration) => {
    let integr = integration;
    integr.applicationID = this.props.match.params.applicationID;

    ApplicationStore.createThingsBoardIntegration(integr, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
    });
  } 

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardHeader title="Add ThingsBoard integration" />
            <CardContent>
              <ThingsBoardIntegrationForm submitLabel="Add integration" onSubmit={this.onSubmit} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default CreateThingsBoardIntegration;
