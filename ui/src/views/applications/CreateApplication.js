import React, { Component } from "react";
import { Link } from "react-router";

import OrganizationSelect from "../../components/OrganizationSelect";
import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationForm from "../../components/ApplicationForm";

class CreateApplication extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      application: {},
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(application) {
    ApplicationStore.createApplication(application, (responseData) => {
      this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+responseData.id);
    });
  }

  componentWillMount() {
    this.setState({
      application: {organizationID: this.props.params.organizationID},
    });
  } 

  render() {
    return (
      <div>
        <ol className="breadcrumb">
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}`}>Dashboard</Link></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications`}>Applications</Link></li>
          <li className="active">Create application</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <ApplicationForm application={this.state.application} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateApplication;
