import React, { Component } from 'react';
import { Link } from 'react-router';

import dispatcher from "../dispatcher";
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
      userDropdownOpen: false,
    }

    this.userToggleDropdown = this.userToggleDropdown.bind(this);
    this.handleActions = this.handleActions.bind(this);
  }

  userToggleDropdown() {
	    this.setState({
	      userDropdownOpen: !this.state.userDropdownOpen,
	    });
	  }

  handleActions(action) {
    switch(action.type) {
      case "BODY_CLICK": {
        this.setState({
            userDropdownOpen: false,
        });
        break;
      }
      default:
        break;
    }
  }

  componentWillMount() {
    SessionStore.on("change", () => {
      this.setState({
        user: SessionStore.getUser(),
        isAdmin: SessionStore.isAdmin(),
      });
    });

    dispatcher.register(this.handleActions);
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
              <li className={typeof(this.state.user.username) === "undefined" ? "hidden" : ""}><Link to="organizations">Organizations</Link></li>
              <li className={this.state.isAdmin === true ? "" : "hidden"}><Link to="users">Users</Link></li>
              <li className={"dropdown " + (typeof(this.state.user.username) === "undefined" ? "hidden" : "") + (this.state.userDropdownOpen ? "open" : "")}>
                <Link onClick={this.userToggleDropdown} className="dropdown-toggle">{this.state.user.username} <span className="caret" /></Link>
                <ul className="dropdown-menu" onClick={this.userToggleDropdown}>
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
