import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";

import ApplicationStore from "../../stores/ApplicationStore";
import IntegrationForm from "./IntegrationForm";


const styles = {
  card: {
    overflow: "visible",
  },
};


class CreateIntegration extends Component {
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
  }

  onSubmit(integration) {
    let integr = integration;
    integr.applicationID = this.props.match.params.applicationID;

    switch (integr.kind) {
      case "http":
        ApplicationStore.createHTTPIntegration(integr, resp => {
          this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
        });
        break;
      case "influxdb":
        ApplicationStore.createInfluxDBIntegration(integr, resp => {
          this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
        });
        break;
      case "thingsboard":
        ApplicationStore.createThingsBoardIntegration(integr, resp => {
          this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
        });
        break;
      default:
        break;
    }
  }

  render() {
    if (this.state.application === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Applications" to={`/organizations/${this.props.match.params.organizationID}/applications`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.application.application.name} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Integrations" to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Create" />
        </TitleBar>
        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <IntegrationForm
                submitLabel="Create integration"
                onSubmit={this.onSubmit}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(CreateIntegration));
