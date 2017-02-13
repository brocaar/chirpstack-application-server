import React, { Component } from 'react';
import { Link } from 'react-router';

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
      this.context.router.push('/applications/'+application.id);
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li><Link to={`/applications/${this.props.params.applicationID}`}>{this.state.application.name}</Link></li>
          <li className="active">Edit application</li>
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

export default UpdateApplication;
