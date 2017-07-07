import React, { Component } from 'react';

class PasswordForm extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      password: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  onChange(field, e) {
    let password = this.state.password;
    password[field] = e.target.value;
    this.setState({
      password: password,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.password);
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="password">Password</label>
          <input className="form-control" id="password" type="password" placeholder="password" value={this.state.password.password || ''} onChange={this.onChange.bind(this, 'password')} />
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.context.router.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

export default PasswordForm;
