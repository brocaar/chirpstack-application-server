import React, { Component } from 'react';

class UserForm extends Component {
  constructor() {
    super();

    this.state = {
      user: {},  
      showPasswordField: true,
    }

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentWillMount() {
    this.setState({
      showPasswordField: (typeof(this.props.user.id) === "undefined"),
      user: this.props.user,
    });
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      showPasswordField: (typeof(nextProps.user.id) === "undefined"),
      user: nextProps.user,
    });
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

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.user);
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="username">Username</label>
          <input className="form-control" id="username" type="text" placeholder="username" required value={this.state.user.username || ''} onChange={this.onChange.bind(this, 'username')} />
        </div>
        <div className={"form-group " + (this.state.showPasswordField ? '' : 'hidden')}>
          <label className="control-label" htmlFor="password">Password</label>
          <input className="form-control" id="password" type="password" placeholder="password" value={this.state.user.password || ''} onChange={this.onChange.bind(this, 'password')} />
        </div>
        <div className="form-group">
          <label className="checkbox-inline">
            <input type="checkbox" name="isActive" id="isActive" checked={this.state.user.isActive} onChange={this.onChange.bind(this, 'isActive')} /> Is active &nbsp;
          </label>
          <label className="checkbox-inline">
            <input type="checkbox" name="isAdmin" id="isAdmin" checked={this.state.user.isAdmin} onChange={this.onChange.bind(this, 'isAdmin')} /> Is global admin &nbsp;
          </label>
        </div>
        <hr />
        <button type="submit" className="btn btn-primary pull-right">Submit</button>
      </form>
    );
  }
}

export default UserForm;
