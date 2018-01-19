import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import Select from "react-select";

import OrganizationStore from "../../stores/OrganizationStore";
import UserStore from "../../stores/UserStore";
import SessionStore from "../../stores/SessionStore";


class AssignUserForm extends Component {
  constructor() {
    super();

    this.state = {
      user: {},
      initialOptions: [],
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.onAutocompleteSelect = this.onAutocompleteSelect.bind(this);
    this.onAutocomplete = this.onAutocomplete.bind(this);
    this.setInitialOptions = this.setInitialOptions.bind(this);
  }

  setInitialOptions() {
    if (this.state.initialOptions.length === 0) {
      UserStore.getAll("", 10, 0, (totalCount, users) => {
        const options = users.map((user, i) => {
          return {
            value: user.id,
            label: user.username,
          };
        });

        this.setState({
          initialOptions: options,
        });
      });
    }
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
    this.setState({user: user});
  }

  onAutocompleteSelect(val) {
    let user = this.state.user;
    user.userID = val.value;
    this.setState({user: user});
  }

  onAutocomplete(input, callbackFunc) {
    UserStore.getAll(input, 10, 0, (totalCount, users) => {
      const options = users.map((user, i) => {
        return {
          value: user.id,
          label: user.username,
      }}); 

      callbackFunc(null, {
        options: options,
        complete: true,
      });
    }); 
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Username</label>
          <Select.Async name="username" required onOpen={this.setInitialOptions} options={this.state.initialOptions} loadOptions={this.onAutocomplete} value={this.state.user.userID} onChange={this.onAutocompleteSelect} clearable={false} autoload={false} />
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


class CreateUserForm extends Component {
  constructor() {
    super();

    this.state = {
      user: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
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
    this.setState({user: user});
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="username">Username</label>
          <input className="form-control" id="username" type="text" placeholder="username" required value={this.state.user.username || ''} onChange={this.onChange.bind(this, 'username')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="email">E-mail address</label>
          <input className="form-control" id="email" type="email" placeholder="e-mail address" required value={this.state.user.email || ''} onChange={this.onChange.bind(this, 'email')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="note">Optional note</label>
          <input className="form-control" id="note" type="text" placeholder="optional note (e.g. phone, address, ...)" value={this.state.user.note || ''} onChange={this.onChange.bind(this, 'note')} />
          <p className="help-block">
            Optional note, e.g. a phone number, address, comment...
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="password">Password</label>
          <input className="form-control" id="password" type="password" placeholder="password" value={this.state.user.password || ''} onChange={this.onChange.bind(this, 'password')} />
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


class CreateOrganizationUser extends Component {
  constructor() {
    super();

    this.state = {
      activeTab: "create",
      displayAssignUser: false,
    };

    this.changeTab = this.changeTab.bind(this);
    this.handleAssign = this.handleAssign.bind(this);
    this.handleCreateAndAssign = this.handleCreateAndAssign.bind(this);
  }

  componentDidMount() {
    this.setState({
      displayAssignUser: SessionStore.isAdmin() || !SessionStore.getSetting('disableAssignExistingUsers'),
      activeTab: (SessionStore.isAdmin() || !SessionStore.getSetting('disableAssignExistingUsers')) ? 'assign' : 'create',
    });

    SessionStore.on("change", () => {
      this.setState({
        displayAssignUser: SessionStore.isAdmin() || !SessionStore.getSetting('disableAssignExistingUsers'),
        activeTab: (SessionStore.isAdmin() || !SessionStore.getSetting('disableAssignExistingUsers')) ? 'assign' : 'create',
      });
    });
  }

  changeTab(e) {
    e.preventDefault();
    this.setState({
      activeTab: e.target.getAttribute('aria-controls'),
    });
  }

  handleAssign(user) {
    OrganizationStore.addUser(this.props.match.params.organizationID, user, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
    }); 
  }

  handleCreateAndAssign(user) {
    UserStore.createUser({username: user.username, email: user.email, note: user.note, password: user.password, isActive: true, organizations: [{organizationID: this.props.match.params.organizationID, isAdmin: user.isAdmin}]}, (resp) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <ul className="nav nav-tabs">
            <li role="presentation" className={(this.state.activeTab === "assign" ? 'active' : '') + " " + (this.state.displayAssignUser ? '' : 'hidden')}><a onClick={this.changeTab} href="#assign" aria-controls="assign">Assign existing user</a></li>
            <li role="presentation" className={(this.state.activeTab === "create" ? 'active' : '')}><a onClick={this.changeTab} href="#create" aria-controls="create">Create and assign user</a></li>
          </ul>
          <hr />
          <div className={(this.state.activeTab === "assign" ? '' : 'hidden')}>
            <AssignUserForm history={this.props.history} onSubmit={this.handleAssign} />
          </div>
          <div className={(this.state.activeTab === "create" ? '' : 'hidden')}>
            <CreateUserForm history={this.props.history} onSubmit={this.handleCreateAndAssign} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(CreateOrganizationUser);
