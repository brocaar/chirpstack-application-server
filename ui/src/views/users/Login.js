import React, { Component } from 'react';
import { withRouter } from "react-router-dom";
import SessionStore from "../../stores/SessionStore";

class Login extends Component {
  constructor() {
    super();

    this.state = {
      login: {},
      registration: null,
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    SessionStore.logout(() => {});
    this.setState({
      registration: SessionStore.getRegistration(),
    });

    SessionStore.on("change", () => {
      this.setState({
        registration: SessionStore.getRegistration(),
      });
    })
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
      this.props.history.push("/");
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
            <div dangerouslySetInnerHTML={{ __html: (typeof(this.state.registration) === "undefined" ? "" : this.state.registration) }}/>
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(Login);
