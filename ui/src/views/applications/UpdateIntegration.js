import React, { Component } from 'react';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";

import ApplicationStore from "../../stores/ApplicationStore";
import IntegrationForm from "./IntegrationForm";


class UpdateIntegration extends Component {
  constructor() {
    super();
    this.state = {};
    this.onSubmit = this.onSubmit.bind(this);
    this.deleteIntegration = this.deleteIntegration.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });

    switch (this.props.match.params.kind) {
      case "http":
        ApplicationStore.getHTTPIntegration(this.props.match.params.applicationID, resp => {
          let integration = resp.integration;
          integration.kind = "http";

          this.setState({
            integration: integration,
          });
        });
        break;
      case "influxdb":
        ApplicationStore.getInfluxDBIntegration(this.props.match.params.applicationID, resp => {
          let integration = resp.integration;
          integration.kind = "influxdb";

          this.setState({
            integration: integration,
          });
        });
        break;
      default:
        break;
    }
  }

  onSubmit(integration) {
    switch (integration.kind) {
      case "http":
        ApplicationStore.updateHTTPIntegration(integration, resp => {
          this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
        });
        break;
      case "influxdb":
        ApplicationStore.updateInfluxDBIntegration(integration, resp => {
          this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
        });
        break;
      default:
        break;
    }
  }

  deleteIntegration() {
    if (window.confirm("Are you sure you want to delete this integration?")) {
      switch(this.props.match.params.kind) {
        case "http":
          ApplicationStore.deleteHTTPIntegration(this.props.match.params.applicationID, resp => {
            this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
          });
          break;
        case "influxdb":
          ApplicationStore.deleteInfluxDBIntegration(this.props.match.params.applicationID, resp => {
            this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
          });
          break;
        default:
          break;
      }
    }
  }

  render() {
    if (this.state.application === undefined || this.state.integration === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={24}>
        <TitleBar
          buttons={[
            <TitleBarButton
              key={1}
              label="Delete"
              onClick={this.deleteIntegration}
            />,
          ]}
        >
          <TitleBarTitle title="Applications" to={`/organizations/${this.props.match.params.organizationID}/applications`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.application.application.name} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Integrations" to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.props.match.params.kind} />
        </TitleBar>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <IntegrationForm
                submitLabel="Update integration"
                object={this.state.integration}
                onSubmit={this.onSubmit}
                update={true}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default UpdateIntegration;
