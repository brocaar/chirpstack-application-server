import React, { Component } from 'react';
import { Link } from 'react-router';

import UserStore from "../../stores/UserStore";
import UserForm from "../../components/UserForm";

class UpdateUser extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      user: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    UserStore.getUser(this.props.params.userID, (user) => {
      this.setState({
        user: user,
      });
    });
  }

  onSubmit(user) {
    UserStore.updateUser(this.props.params.userID, this.state.user, (responseData) => {
      this.context.router.push('/users');
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this user?")) {
      UserStore.deleteUser(this.props.params.userID, (responseData) => {
        this.context.router.push('/users/');
      });
    } 
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/users">Users</Link></li>
          <li className="active">{this.state.user.username}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/users/${this.props.params.userID}/password`}><button type="button" className="btn btn-default">Change password</button></Link> &nbsp;
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete user</button></Link>
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

export default UpdateUser;
