import React, { Component } from 'react';
import { Link } from 'react-router';

import tokenStore from "../../stores/TokenStore";

class JWTToken extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      token: tokenStore.getToken(),
    };

    this.onChange = this.onChange.bind(this);
    this.onSubmit = this.onSubmit.bind(this);
  }

  onChange(e) {
    this.setState({token: e.target.value});
  }

  onSubmit(e) {
    e.preventDefault();
    tokenStore.setToken(this.state.token);
    this.context.router.push("/");
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li className="active">JWT Token</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <form onSubmit={this.onSubmit}>
              <div className="form-group">
                <label className="control-label" htmlFor="name">JWT Token</label>
                <textarea className="form-control" rows="5" id="token" name="token" onChange={this.onChange}>{this.state.token}</textarea>
                <p className="help-block">
                  LoRa App Server has support for JWT based token authorization as described in the <a href="https://docs.loraserver.io/lora-app-server/api/">api documentation</a>.
                  When you got redirected to this view, it means you need to enter your token to gain access to your data.
                </p>
              </div>
              <hr />
              <button type="submit" className="btn btn-primary pull-right">Set JWT Token</button>
            </form>
          </div>
        </div>
      </div>
    )
  }
}

export default JWTToken;
