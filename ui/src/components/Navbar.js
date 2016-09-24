import React, { Component } from 'react';

class Navbar extends Component {
  render() {
    return (
      <nav className="navbar navbar-inverse navbar-fixed-top">
        <div className="container-fluid">
          <div className="navbar-header">
            <a className="navbar-brand" href="#">LoRa Server</a>
          </div>
          <div id="navbar" className="navbar-collapse collapse">
            <ul className="nav navbar-nav navbar-right">
              <li><a href="#">Set JWT token</a></li>
              <li><a href="https://github.com/brocaar/loraserver">GitHub</a></li>
              <li><a href="https://docs.loraserver.io/">Documentation</a></li>
            </ul>
          </div>
        </div>
      </nav>
    );
  }
}

export default Navbar;
