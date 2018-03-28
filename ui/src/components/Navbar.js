import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';

import dispatcher from "../dispatcher";
import SessionStore from "../stores/SessionStore";


class Navbar extends Component {
  constructor() {
    super();
    this.state = {
      user: SessionStore.getUser(),
      isAdmin: SessionStore.isAdmin(),
      userDropdownOpen: false,
      logo: SessionStore.getLogo(),
      searching: false,
      search: '',
    }

    this.userToggleDropdown = this.userToggleDropdown.bind(this);
    this.handleActions = this.handleActions.bind(this);
    this.toggleSearch = this.toggleSearch.bind(this);
    this.onSearchSubmit = this.onSearchSubmit.bind(this);
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
        logo: SessionStore.getLogo(),
      });
    });

    dispatcher.register(this.handleActions);
  }

  toggleSearch(e) {
    e.preventDefault();
    this.setState({
      searching: !this.state.searching,
    });
  }

  onSearchChange(field, e) {
    this.setState({
      search: e.target.value,
    });
  }

  onSearchSubmit(e) {
    e.preventDefault();
    this.props.history.push(`/search?search=${encodeURIComponent(this.state.search)}`);
  }

  render() {
    return (
      <nav className="navbar navbar-inverse navbar-fixed-top">
        <div className={`container ${this.state.searching ? '' : 'hidden'}`}>
          <div id="navbar" className="navbar-collapse collapse">
            <form className="navbar-form form-inline" onSubmit={this.onSearchSubmit}>
              <div className="form-group">
                <div className="input-group">
                  <input className="form-control" type="text" placeholder="Search organization, application, gateway or device" value={this.state.search} onChange={this.onSearchChange.bind(this, 'search')} />
                  <span className="input-group-addon">
                    <a href="#search" onClick={this.toggleSearch}><span className="glyphicon glyphicon-remove" aria-hidden="true"></span><span className="hidden">close search</span></a>
                  </span>
                </div>
              </div>
            </form>
          </div>
        </div>
        <div className={`container ${this.state.searching ? 'hidden' : ''}`}>
          <div className="navbar-header">
            <a className="navbar-brand" href="#/">
              <span dangerouslySetInnerHTML={{ __html: ( typeof(this.state.logo) === "undefined" ? "" : this.state.logo) }} />
              LoRa Server
            </a>
          </div>
          <div id="navbar" className="navbar-collapse collapse">
            <ul className="nav navbar-nav navbar-right">
              <li className={typeof(this.state.user.username) === "undefined" ? "hidden" : ""}><Link to="/organizations">Organizations</Link></li>
              <li className={this.state.isAdmin === true ? "" : "hidden"}><Link to="/users">Users</Link></li>
              <li className={this.state.isAdmin === true ? "" : "hidden"}><Link to="/network-servers">Network servers</Link></li>
              <li className={"dropdown " + (typeof(this.state.user.username) === "undefined" ? "hidden" : "") + (this.state.userDropdownOpen ? "open" : "")}>
                <a onClick={this.userToggleDropdown} className="dropdown-toggle">{this.state.user.username} <span className="caret" /></a>
                <ul className="dropdown-menu" onClick={this.userToggleDropdown}>
                  <li><Link to={`/users/${this.state.user.id}/password`}>Change password</Link></li>
                  <li><Link to="/login">Logout</Link></li>
                </ul>
              </li>
              <li className={this.state.user.username === undefined ? "hidden": ""}><a href="#search" onClick={this.toggleSearch}><span className="glyphicon glyphicon-search" aria-hidden="true"></span><span className="hidden">search</span></a></li>
            </ul>
          </div>
        </div>
      </nav>
    );
  }
}

export default withRouter(Navbar);
