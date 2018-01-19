import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationForm from "../../components/ApplicationForm";

class CreateApplication extends Component {
  constructor() {
    super();
    this.state = {
      application: {},
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(application) {
    ApplicationStore.createApplication(application, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${responseData.id}`);
    });
  }

  componentDidMount() {
    this.setState({
      application: {organizationID: this.props.match.params.organizationID},
    });
  } 

  render() {
    return (
      <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title panel-title-buttons">Create application</h3>
        </div>
        <div className="panel-body">
          <ApplicationForm application={this.state.application} onSubmit={this.onSubmit} organizationID={this.props.match.params.organizationID} />
        </div>
      </div>
    );
  }
}

export default withRouter(CreateApplication);
