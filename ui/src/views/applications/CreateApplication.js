import React, { Component } from "react";
import { Link } from "react-router";

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
      this.context.router.push('/applications/'+responseData.id);
    });
  }

  render() {
    return (
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
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
