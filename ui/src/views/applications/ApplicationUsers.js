import React, { Component } from 'react';
import { Link } from 'react-router';

import ApplicationStore from "../../stores/ApplicationStore";
import Pagination from "../../components/Pagination";

class ApplicationUserRow extends Component {
  render() {
    return(
      <tr>
        <td>{this.props.user.id}</td>
        <td>
          <Link to={`applications/${this.props.application.id}/users/${this.props.user.id}/edit`}>{this.props.user.username}</Link>
        </td>
        <td>
          <span className={"glyphicon glyphicon-" + (this.props.user.isAdmin ? 'ok' : 'remove')} aria-hidden="true"></span>
        </td>
      </tr>    
    );
  }
}


class ApplicationUsers extends Component {
  constructor() {
    super();

    this.state = {
      application: {},
      users: [],
      pageSize: 20,
      pageNumber: 1,
      pages: 1,
    };

    this.updatePage = this.updatePage.bind(this);
  }

  componentDidMount() {
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({
        application: application,
      });
    });

    this.updatePage(this.props);
  }

  componentWillReceiveProps(nextProps) {
    this.updatePage(nextProps);
  }

  updatePage(props) {
    const page = (props.location.query.page === undefined) ? 1 : props.location.query.page;

    ApplicationStore.getUsers(this.props.params.applicationID, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, users) => {
      this.setState({
        users: users,
        pages: Math.ceil(totalCount / this.state.pageSize),
        pageNumber: page,
      });
    });
  }

  render() {
    const UserRows = this.state.users.map((user, i) => <ApplicationUserRow key={user.id} application={this.state.application} user={user} />);

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li><Link to={`/applications/${this.state.application.id}`}>{this.state.application.name}</Link></li>
          <li className="active">Users</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/applications/${this.state.application.id}/users/create`}><button type="button" className="btn btn-default">Add user</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th className="col-md-1">ID</th>
                  <th>Username</th>
                  <th className="col-md-1">Admin</th>
                </tr>
              </thead>
              <tbody>
                {UserRows}
              </tbody>
            </table>
          </div>
          <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`applications/${this.state.application.id}/users`} />
        </div>
      </div>
    );
  }
}

export default ApplicationUsers;
