import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationIntegrationForm from "../../components/ApplicationIntegrationForm";


class CreateApplicationIntegration extends Component {
  constructor() {
    super();

    this.state = {
      integration: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(integration) {
    ApplicationStore.createHTTPIntegration(this.props.match.params.applicationID, integration, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title panel-title-buttons">Add integration</h3>
        </div>
        <div className="panel-body">
          <ApplicationIntegrationForm integration={this.state.integration} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(CreateApplicationIntegration);
