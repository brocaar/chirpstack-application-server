import React, { Component } from 'react';

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationForm from "../../components/ApplicationForm";

class UpdateApplication extends Component {
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

  componentDidMount() {
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({application: application});
    });
  }

  onSubmit(application) {
    ApplicationStore.updateApplication(this.props.params.applicationID, this.state.application, (responseData) => {
      this.context.router.push('/organizations/'+application.organizationID+'/applications/'+application.id);
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <ApplicationForm application={this.state.application} onSubmit={this.onSubmit} update={true} organizationID={this.props.params.organizationID} />
        </div>
      </div>
    );
  }
}

export default UpdateApplication;
