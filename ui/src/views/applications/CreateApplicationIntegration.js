import React, { Component } from 'react';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationIntegrationForm from "../../components/ApplicationIntegrationForm";


class CreateApplicationIntegration extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      integration: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(integration) {
    ApplicationStore.createHTTPIntegration(this.props.params.applicationID, integration, (responseData) => {
      this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+this.props.params.applicationID+'/integrations');
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

export default CreateApplicationIntegration;
