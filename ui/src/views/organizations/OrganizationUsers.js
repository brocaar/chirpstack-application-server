import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import OrganizationStore from "../../stores/OrganizationStore";
import Pagination from "../../components/Pagination";


class OrganizationUserRow extends Component {
  render() {
    return(
      <tr>
        <td>{this.props.user.id}</td>
        <td>
          <Link to={`/organizations/${this.props.organizationID}/users/${this.props.user.id}/edit`}>{this.props.user.username}</Link>
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
    this.updatePage(this.props);
  }

  componentWillReceiveProps(nextProps) {
    this.updatePage(nextProps);
  }

  updatePage(props) {
    const query = new URLSearchParams(props.location.search);
    const page = (query.get('page') === null) ? 1 : query.get('page');

    OrganizationStore.getUsers(this.props.match.params.organizationID, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, users) => {
      this.setState({
        users: users,
        pages: Math.ceil(totalCount / this.state.pageSize),
        pageNumber: page,
      });
    });
  }

  render() {
    const UserRows = this.state.users.map((user, i) => <OrganizationUserRow key={user.id} organizationID={this.props.match.params.organizationID} user={user} />);

    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <div className="btn-group pull-right">
           <Link to={`/organizations/${this.props.match.params.organizationID}/users/create`}><button type="button" className="btn btn-default btn-sm">Add user</button></Link>
          </div>
        </div>
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
        <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.match.params.organizationID}/users`} />
      </div>
    );
  }
}

export default OrganizationUsers;
