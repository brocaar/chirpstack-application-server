import React, { Component } from 'react';
import { Link } from 'react-router';

import ApplicationStore from "../../stores/ApplicationStore";

class UpdateApplicationUserForm extends Component {
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
              <input type="checkbox" name="isAdmin" id="isAdmin" checked={this.state.user.isAdmin} onChange={this.onChange.bind(this, 'isAdmin')} /> Is application admin
            </label>
          </div>
          <p className="help-block">
            When checked, the user will be assigned admin permissions within the context of the application.
          </p>
        </div>
        <hr />
        <button type="submit" className="btn btn-primary pull-right">Submit</button>
      </form>
    );
  }
}

class UpdateApplicationUser extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      application: {},
      user: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({
        application: application,
      });
    });

    ApplicationStore.getUser(this.props.params.applicationID, this.props.params.userID, (user) => {
      this.setState({
        user: user,
      });
    });
  }

  onSubmit(user) {
    ApplicationStore.updateUser(this.props.params.applicationID, this.props.params.userID, user, (responseData) => {
      this.context.router.push("/applications/"+this.props.params.applicationID+"/users");
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this application user (this does not remove the user itself)?")) {
      ApplicationStore.removeUser(this.props.params.applicationID, this.props.params.userID, (responseData) => {
        this.context.router.push("/applications/"+this.props.params.applicationID+"/users");
      }); 
    }
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li><Link to={`/applications/${this.state.application.id}`}>{this.state.application.name}</Link></li>
          <li><Link to={`/applications/${this.state.application.id}/users`}>Users</Link></li>
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
            <UpdateApplicationUserForm user={this.state.user} onSubmit={this.onSubmit}  />
          </div>
        </div>
      </div>
    );
  }
}

export default UpdateApplicationUser;
