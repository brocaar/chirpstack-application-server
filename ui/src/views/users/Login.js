import React, { Component } from 'react';
import SessionStore from "../../stores/SessionStore";

class Login extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      login: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    SessionStore.logout(() => {});
  }

  onChange(field, e) {
    let login = this.state.login;
    login[field] = e.target.value;
    this.setState({
      login: login,
    });
  }

  onSubmit(e) {
    e.preventDefault(); 
    SessionStore.login(this.state.login, (token) => {
      this.context.router.push("/");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li className="active">Login</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <form onSubmit={this.onSubmit}>
              <div className="form-group">
                <label className="control-label" htmlFor="username">Username</label>
                <input className="form-control" id="username" type="text" placeholder="username" required value={this.state.login.username || ''} onChange={this.onChange.bind(this, 'username')} />
              </div>
              <div className="form-group">
                <label className="control-label" htmlFor="password">Password</label>
                <input className="form-control" id="password" type="password" placeholder="password" value={this.state.login.password || ''} onChange={this.onChange.bind(this, 'password')} />
              </div>
              <hr />
              <button type="submit" className="btn btn-primary pull-right">Login</button>
            </form>
          </div>
        </div>
      </div>
    );
  }
}

export default Login;
