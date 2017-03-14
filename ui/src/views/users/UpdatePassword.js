import React, { Component } from 'react';
import { Link } from 'react-router';

import UserStore from "../../stores/UserStore";
import PasswordForm from "../../components/PasswordForm";

class UpdatePassword extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      user: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    UserStore.getUser(this.props.params.userID, (user) => {
      this.setState({
        user: user,
      });
    });
  }

  onSubmit(password) {
    UserStore.updatePassword(this.props.params.userID, password, (responseData) => {
      if (this.state.user.isAdmin) {
        this.context.router.push('/users/' + this.props.params.userID + "/edit");
      } else {
        // non-admin users don't have access to /users view
        this.context.router.push("/");
      }
    });    
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/users">Users</Link></li>
          <li><Link to={`/users/${this.props.params.userID}/edit`}>{this.state.user.username}</Link></li>
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

export default UpdatePassword;
