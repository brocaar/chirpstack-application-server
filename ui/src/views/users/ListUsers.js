import React, { Component } from 'react';
import { Link } from 'react-router';

import Pagination from "../../components/Pagination";
import UserStore from "../../stores/UserStore";

class UserRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/users/${this.props.user.id}/edit`}>{this.props.user.username}</Link></td>
        <td><span className={"glyphicon glyphicon-" + (this.props.user.isActive ? 'ok' : 'remove')} aria-hidden="true"></span></td>
        <td><span className={"glyphicon glyphicon-" + (this.props.user.isAdmin ? 'ok' : 'remove')} aria-hidden="true"></span></td>
      </tr>
    );
  }
}

class ListUsers extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
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
    const page = (props.location.query.page === undefined) ? 1 : props.location.query.page;

    UserStore.getAll("", this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, users) => {
      this.setState({
        users: users,
        pageNumber: page,
        pages: Math.ceil(totalCount / this.state.pageSize),
      });
      window.scrollTo(0, 0);
    });
  }

  render() {
    const UserRows = this.state.users.map((user, i) => <UserRow key={user.id} user={user} />);

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li className="active">Users</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to="/users/create"><button type="button" className="btn btn-default">Create user</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>Username</th>
                  <th className="col-md-1">Active</th>
                  <th className="col-md-1">Admin</th>
                </tr>
              </thead>
              <tbody>
                {UserRows}
              </tbody>
            </table>
          </div>
          <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname="/users" />
        </div>
      </div>
    );
  }
}

export default ListUsers;
