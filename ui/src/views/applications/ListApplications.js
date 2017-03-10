import React, { Component } from 'react';
import { Link } from 'react-router';

import ApplicationStore from "../../stores/ApplicationStore";

class ApplicationRow extends Component {
  render() {
    return(
      <tr>
        <td>{this.props.application.id}</td>
        <td><Link to={`/applications/${this.props.application.id}`}>{this.props.application.name}</Link></td>
        <td>{this.props.application.description}</td>
      </tr>
    );
  }
}

class ListApplications extends Component {
  constructor() {
    super();
    this.state = {
      applications: [],
    };
    ApplicationStore.getAll((applications) => {
      this.setState({applications: applications});
    });
  }

  render () {
    const ApplicationRows = this.state.applications.map((application, i) => <ApplicationRow key={application.id} application={application} />);

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li className="active">Applications</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to="/applications/create"><button type="button" className="btn btn-default">Create application</button></Link> &nbsp;
            <Link to="/channels"><button type="button" className="btn btn-default">Channel lists</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th className="col-md-1">ID</th>
                  <th className="col-md-4">Name</th>
                  <th>Description</th>
                </tr>
              </thead>
              <tbody>
                {ApplicationRows}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}

export default ListApplications;
