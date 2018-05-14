import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationIntegrationForm from "../../components/ApplicationIntegrationForm";


class UpdateApplicationIntegration extends Component {
  constructor() {
    super();

    this.state = {
      integration: {
        configuration: {},
      },
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    ApplicationStore.getIntegration(this.props.match.params.applicationID, this.props.match.params.kind, (integration) => {
      integration.kind = this.props.match.params.kind;
      this.setState({
        integration: integration,
      }); 
    });
  }

  onSubmit(integration) {
    ApplicationStore.updateIntegration(this.props.match.params.applicationID, integration.kind, integration, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`);
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this integration?")) {
      ApplicationStore.deleteIntegration(this.props.match.params.applicationID, this.state.integration.kind, (responseData) => {
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
