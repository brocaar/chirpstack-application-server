import React, { Component } from 'react';
import { Link } from 'react-router';

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
    };

  }

  componentWillMount() {
    UserStore.getAll("", (users) => {
      this.setState({users: users});
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
        </div>
      </div>
    );
  }
}

export default ListUsers;
