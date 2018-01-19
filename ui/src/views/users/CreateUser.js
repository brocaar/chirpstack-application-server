import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';

import UserStore from "../../stores/UserStore";
import UserForm from "../../components/UserForm";

class CreateUser extends Component {
  constructor() {
    super();

    this.state = {
      user: {
        isActive: true,
      },
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(user) {
    UserStore.createUser(user, (responseData) => {
      this.props.history.push('/users');
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/users">Users</Link></li>
          <li className="active">Create user</li>
        </ol>
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

export default withRouter(CreateUser);
