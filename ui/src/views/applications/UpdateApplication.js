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

  componentWillMount() {
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
          <ApplicationForm application={this.state.application} onSubmit={this.onSubmit} update={true} />
        </div>
      </div>
    );
  }
}

export default UpdateApplication;
