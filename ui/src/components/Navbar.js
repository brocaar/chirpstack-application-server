import React, { Component } from 'react';
import { Link } from 'react-router';

import SessionStore from "../stores/SessionStore";

class Navbar extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      user: SessionStore.getUser(),
      isAdmin: SessionStore.isAdmin(),
      dropdownOpen: false,
    }

    this.toggleDropdown = this.toggleDropdown.bind(this);
  }

  toggleDropdown() {
    this.setState({
      dropdownOpen: !this.state.dropdownOpen,
    });
  }

  componentWillMount() {
    SessionStore.on("change", () => {
      this.setState({
        user: SessionStore.getUser(),
        isAdmin: SessionStore.isAdmin(),
      });
    });
  }

  render() {
    return (
      <nav className="navbar navbar-inverse navbar-fixed-top">
        <div className="container">
          <div className="navbar-header">
            <a className="navbar-brand" href="#">LoRa Server</a>
          </div>
          <div id="navbar" className="navbar-collapse collapse">
            <ul className="nav navbar-nav navbar-right">
              <li className={this.state.isAdmin === true ? "" : "hidden"}><Link to="users">Manage users</Link></li>
              <li className={"dropdown " + (typeof(this.state.user.username) === "undefined" ? "hidden" : "") + (this.state.dropdownOpen ? "open" : "")}>
                <Link onClick={this.toggleDropdown} className="dropdown-toggle">{this.state.user.username} <span className="caret" /></Link>
                <ul className="dropdown-menu" onClick={this.toggleDropdown}>
                  <li><Link to={`users/${this.state.user.id}/password`}>Change password</Link></li>
                  <li><Link to="login">Logout</Link></li>
                </ul>
              </li>
            </ul>
          </div>
        </div>
      </nav>
    );
  }
}

export default Navbar;
