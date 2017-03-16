import React, { Component } from 'react';
import { Link } from 'react-router';

import Select from "react-select";

import ApplicationStore from "../../stores/ApplicationStore";
import UserStore from "../../stores/UserStore";

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
  }

  componentWillMount() {
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
          <Select.Async name="username" required options={this.state.initialOptions} loadOptions={this.onAutocomplete} value={this.state.user.userID} onChange={this.onAutocompleteSelect} clearable={false} autoload={false} />
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
          <label className="control-label" htmlFor="password">Password</label>
          <input className="form-control" id="password" type="password" placeholder="password" value={this.state.user.password || ''} onChange={this.onChange.bind(this, 'password')} />
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


class CreateApplicationUser extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      application: {},
      user: {},
      activeTab: "assign",
    };

    this.changeTab = this.changeTab.bind(this);
    this.handleAssign = this.handleAssign.bind(this);
    this.handleCreateAndAssign = this.handleCreateAndAssign.bind(this);
  }

  componentWillMount() {
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({
        application: application,
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
    ApplicationStore.addUser(this.props.params.applicationID, user, (responseData) => {
      this.context.router.push("/applications/"+this.props.params.applicationID+"/users");
    }); 
  }

  handleCreateAndAssign(user) {
    UserStore.createUser({username: user.username, password: user.password, isActive: true}, (resp) => {
      ApplicationStore.addUser(this.props.params.applicationID, {userID: resp.id, isAdmin: user.isAdmin}, (responseData) => {
        this.context.router.push("/applications/"+this.props.params.applicationID+"/users");
      });
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li><Link to={`/applications/${this.state.application.id}`}>{this.state.application.name}</Link></li>
          <li><Link to={`/applications/${this.state.application.id}/users`}>Users</Link></li>
          <li className="active">Add user</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <ul className="nav nav-tabs">
              <li role="presentation" className={(this.state.activeTab === "assign" ? 'active' : '')}><a onClick={this.changeTab} href="#assign" aria-controls="assign">Assign existing user</a></li>
              <li role="presentation" className={(this.state.activeTab === "create" ? 'active' : '')}><a onClick={this.changeTab} href="#create" aria-controls="create">Create and assign user</a></li>
            </ul>
            <hr />
            <div className={(this.state.activeTab === "assign" ? '' : 'hidden')}>
              <AssignUserForm onSubmit={this.handleAssign} />
            </div>
            <div className={(this.state.activeTab === "create" ? '' : 'hidden')}>
              <CreateUserForm onSubmit={this.handleCreateAndAssign} />
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default CreateApplicationUser;
