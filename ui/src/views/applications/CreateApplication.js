import React, { Component } from "react";

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
      <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title panel-title-buttons">Create application</h3>
        </div>
        <div className="panel-body">
          <ApplicationForm application={this.state.application} onSubmit={this.onSubmit} organizationID={this.props.params.organizationID} />
        </div>
      </div>
    );
  }
}

export default CreateApplication;
