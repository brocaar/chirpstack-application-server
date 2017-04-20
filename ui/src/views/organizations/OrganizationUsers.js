import React, { Component } from 'react';
import { Link } from 'react-router';

import OrganizationStore from "../../stores/OrganizationStore";
import Pagination from "../../components/Pagination";

class OrganizationUserRow extends Component {
  render() {
    return(
      <tr>
        <td>{this.props.user.id}</td>
        <td>
          <Link to={`organizations/${this.props.organization.id}/users/${this.props.user.id}/edit`}>{this.props.user.username}</Link>
        </td>
        <td>
          <span className={"glyphicon glyphicon-" + (this.props.user.isAdmin ? 'ok' : 'remove')} aria-hidden="true"></span>
        </td>
      </tr>    
    );
  }
}


class OrganizationUsers extends Component {
  constructor() {
    super();

    this.state = {
      organization: {},
      users: [],
      pageSize: 20,
      pageNumber: 1,
      pages: 1,
    };

    this.updatePage = this.updatePage.bind(this);
  }

  componentDidMount() {
    OrganizationStore.getOrganization(this.props.params.organizationID, (organization) => {
      this.setState({
        organization: organization,
      });
    });

    this.updatePage(this.props);
  }

  componentWillReceiveProps(nextProps) {
    this.updatePage(nextProps);
  }

  updatePage(props) {
    const page = (props.location.query.page === undefined) ? 1 : props.location.query.page;

    OrganizationStore.getUsers(this.props.params.organizationID, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, users) => {
      this.setState({
        users: users,
        pages: Math.ceil(totalCount / this.state.pageSize),
        pageNumber: page,
      });
    });
  }

  render() {
    const UserRows = this.state.users.map((user, i) => <OrganizationUserRow key={user.id} organization={this.state.organization} user={user} />);

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/organizations">Organizations</Link></li>
          <li><Link to={`/organizations/${this.state.organization.id}`}>{this.state.organization.name}</Link></li>
          <li className="active">Users</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/organizations/${this.state.organization.id}/users/create`}><button type="button" className="btn btn-default">Add user</button></Link>
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
          <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`organizations/${this.state.organization.id}/users`} />
        </div>
      </div>
    );
  }
}

export default OrganizationUsers;
