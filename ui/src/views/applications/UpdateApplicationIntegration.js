import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationIntegrationForm from "../../components/ApplicationIntegrationForm";


class UpdateApplicationIntegration extends Component {
  constructor() {
    super();

    this.state = {
      integration: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    ApplicationStore.getHTTPIntegration(this.props.match.params.applicationID, (integration) => {
      integration.kind = "http";
      this.setState({
        integration: integration,
      }); 
    });
  }

  onSubmit(integration) {
    ApplicationStore.updateHTTPIntegration(this.props.match.params.applicationID, integration, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this integration?")) {
      ApplicationStore.deleteHTTPIntegration(this.props.match.params.applicationID, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
      });
    }
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <h3 className="panel-title panel-title-buttons pull-left">Update integration</h3>
          <div className="btn-group pull-right">
            <a><button type="button" className="btn btn-danger btn-sm" onClick={this.onDelete}>Remove integration</button></a>
          </div>
        </div>
        <div className="panel-body">
          <ApplicationIntegrationForm integration={this.state.integration} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateApplicationIntegration);
