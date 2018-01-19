import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';

import UserStore from "../../stores/UserStore";
import UserForm from "../../components/UserForm";

class UpdateUser extends Component {
  constructor() {
    super();

    this.state = {
      user: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    UserStore.getUser(this.props.match.params.userID, (user) => {
      this.setState({
        user: user,
      });
    });
  }

  onSubmit(user) {
    UserStore.updateUser(this.props.match.params.userID, this.state.user, (responseData) => {
      this.props.history.push('/users');
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this user?")) {
      UserStore.deleteUser(this.props.match.params.userID, (responseData) => {
        this.props.history.push('/users/');
      });
    } 
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/users">Users</Link></li>
          <li className="active">{this.state.user.username}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/users/${this.props.match.params.userID}/password`}><button type="button" className="btn btn-default">Change password</button></Link> &nbsp;
            <a><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete user</button></a>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <UserForm user={this.state.user} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateUser);
