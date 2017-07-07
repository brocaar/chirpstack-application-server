import React, { Component } from 'react';
import { Link } from 'react-router';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationIntegrationForm from "../../components/ApplicationIntegrationForm";


class UpdateApplicationIntegration extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      integration: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    ApplicationStore.getHTTPIntegration(this.props.params.applicationID, (integration) => {
      integration.kind = "http";
      this.setState({
        integration: integration,
      }); 
    });
  }

  onSubmit(integration) {
    ApplicationStore.updateHTTPIntegration(this.props.params.applicationID, integration, (responseData) => {
      this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+this.props.params.applicationID+'/integrations');
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this integration?")) {
      ApplicationStore.deleteHTTPIntegration(this.props.params.applicationID, (responseData) => {
        this.context.router.push("/organizations/"+this.props.params.organizationID+"/applications/"+this.props.params.applicationID+"/integrations");
      });
    }
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <h3 className="panel-title panel-title-buttons pull-left">Update integration</h3>
          <div className="btn-group pull-right">
            <Link><button type="button" className="btn btn-danger btn-sm" onClick={this.onDelete}>Remove integration</button></Link>
          </div>
        </div>
        <div className="panel-body">
          <ApplicationIntegrationForm integration={this.state.integration} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default UpdateApplicationIntegration;
