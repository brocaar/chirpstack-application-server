import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import OrganizationStore from "../../stores/OrganizationStore";


class UpdateOrganizationUserForm extends Component {
  constructor() {
    super();

    this.state = {
      user: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      user: nextProps.user,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.user);
  }

  onChange(field, e) {
    let user = this.state.user;
    if (e.target.type === "checkbox") {
      user[field] = e.target.checked;
    } else {
      user[field] = e.target.value;
    }

    this.setState({
      user: user,
    });
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="username">Username</label>
          <input className="form-control" id="username" type="text" placeholder="username" disabled value={this.state.user.username || ''} />
        </div>
        <div className="form-group">
          <label className="control-label">Admin</label>
          <div className="checkbox">
            <label>
              <input type="checkbox" name="isAdmin" id="isAdmin" checked={!!this.state.user.isAdmin} onChange={this.onChange.bind(this, 'isAdmin')} /> Is organization admin
            </label>
          </div>
          <p className="help-block">
            When checked, the user will be assigned admin permissions within the context of the organization.
          </p>
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}


class UpdateOrganizationUser extends Component {
  constructor() {
    super();

    this.state = {
      user: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    OrganizationStore.getUser(this.props.match.params.organizationID, this.props.match.params.userID, (user) => {
      this.setState({
        user: user,
      });
    });    
  }

  onSubmit(user) {
    OrganizationStore.updateUser(this.props.match.params.organizationID, this.props.match.params.userID, user, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this organization user (this does not remove the user itself)?")) {
      OrganizationStore.removeUser(this.props.match.params.organizationID, this.props.match.params.userID, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
      }); 
    }
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <h3 className="panel-title panel-title-buttons pull-left">Update user</h3>
          <div className="btn-group pull-right">
            <a><button type="button" className="btn btn-danger btn-sm" onClick={this.onDelete}>Remove user</button></a>
          </div>
        </div>
        <div className="panel-body">
          <UpdateOrganizationUserForm history={this.props.history} user={this.state.user} onSubmit={this.onSubmit}  />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateOrganizationUser);
