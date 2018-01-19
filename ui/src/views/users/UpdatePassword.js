import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';

import UserStore from "../../stores/UserStore";
import PasswordForm from "../../components/PasswordForm";

class UpdatePassword extends Component {
  constructor() {
    super();

    this.state = {
      user: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    UserStore.getUser(this.props.match.params.userID, (user) => {
      this.setState({
        user: user,
      });
    });
  }

  onSubmit(password) {
    UserStore.updatePassword(this.props.match.params.userID, password, (responseData) => {
      if (this.state.user.isAdmin) {
        this.props.history.push('/users/' + this.props.match.params.userID + "/edit");
      } else {
        // non-admin users don't have access to /users view
        this.props.history.push("/");
      }
    });    
  }

  render() {
    return(
      <div>
        <ol className={"breadcrumb " + (!this.state.user.isAdmin ? 'hidden' : '')}>
          <li><Link to="/users">Users</Link></li>
          <li><Link to={`/users/${this.props.match.params.userID}/edit`}>{this.state.user.username}</Link></li>
          <li className="active">Update password</li>
        </ol>
        <ol className={"breadcrumb " + (this.state.user.isAdmin ? 'hidden' : '')}>
          <li>Users</li>
          <li>{this.state.user.username}</li>
          <li className="active">Update password</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <PasswordForm onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(UpdatePassword);
