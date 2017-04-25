import React, { Component } from 'react';
import { Link } from 'react-router';

import OrganizationSelect from "../../components/OrganizationSelect";
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
              <input type="checkbox" name="isAdmin" id="isAdmin" checked={this.state.user.isAdmin} onChange={this.onChange.bind(this, 'isAdmin')} /> Is organization admin
            </label>
          </div>
          <p className="help-block">
            When checked, the user will be assigned admin permissions within the context of the organization.
          </p>
        </div>
        <hr />
        <button type="submit" className="btn btn-primary pull-right">Submit</button>
      </form>
    );
  }
}


class UpdateOrganizationUser extends Component {
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
    OrganizationStore.getUser(this.props.params.organizationID, this.props.params.userID, (user) => {
      this.setState({
        user: user,
      });
    });    
  }

  onSubmit(user) {
    OrganizationStore.updateUser(this.props.params.organizationID, this.props.params.userID, user, (responseData) => {
      this.context.router.push("/organizations/"+this.props.params.organizationID+"/users");
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this organization user (this does not remove the user itself)?")) {
      OrganizationStore.removeUser(this.props.params.organizationID, this.props.params.userID, (responseData) => {
        this.context.router.push("/organizations/"+this.props.params.organizationID+"/users");
      }); 
    }
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/users`}>Users</Link></li>
          <li className="active">{this.state.user.username}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Remove user</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <UpdateOrganizationUserForm user={this.state.user} onSubmit={this.onSubmit}  />
          </div>
        </div>
      </div>
    );
  }
}

export default UpdateOrganizationUser;
