import React, { Component } from "react";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";

import ApplicationStore from "../../../stores/ApplicationStore";
import AWSSNSIntegrationForm from "./AWSSNSIntegrationForm";


class UpdateAWSSNSIntegration extends Component {
  constructor() {
    super();

    this.state = {};
  }

  onSubmit = (integration) => {
    let integr = integration;
    integr.applicationID = this.props.match.params.applicationID;

    ApplicationStore.updateAWSSNSIntegration(integr, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
    });
  } 

  componentDidMount() {
    ApplicationStore.getAWSSNSIntegration(this.props.match.params.applicationID, (resp) => {
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
            <CardHeader title="Update AWS SNS integration" />
            <CardContent>
              <AWSSNSIntegrationForm submitLabel="Update integration" onSubmit={this.onSubmit} object={this.state.object} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default UpdateAWSSNSIntegration;
