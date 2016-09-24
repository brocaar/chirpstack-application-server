import React, { Component } from 'react';
import Navbar from "./components/Navbar";
import Errors from "./components/Errors";

class Layout extends Component {
  render() {
    return (
      <div>
        <Navbar />
        <div className="container">
          <div className="row">
            <Errors />
            {this.props.children}
          </div>
        </div>
      </div>
    );
  }
}

export default Layout;
