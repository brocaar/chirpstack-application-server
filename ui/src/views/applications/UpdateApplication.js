import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationForm from "../../components/ApplicationForm";


class UpdateApplication extends Component {
  constructor() {
    super();
    this.state = {
      application: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    ApplicationStore.getApplication(this.props.match.params.applicationID, (application) => {
      this.setState({application: application});
    });
  }

  onSubmit(application) {
    ApplicationStore.updateApplication(this.props.match.params.applicationID, this.state.application, (responseData) => {
      this.props.history.push(`/organizations/${application.organizationID}/applications/${application.id}`);
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <ApplicationForm application={this.state.application} onSubmit={this.onSubmit} update={true} organizationID={this.props.match.params.organizationID} />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateApplication);
