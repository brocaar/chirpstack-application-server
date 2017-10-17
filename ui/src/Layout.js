import React, { Component } from 'react';
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import Errors from "./components/Errors";
import dispatcher from "./dispatcher";

class Layout extends Component {
  onClick() {
    dispatcher.dispatch({
      type: "BODY_CLICK",
    });
  }

  render() {
    return (
      <div>
        <Navbar />
        <div className="container" onClick={this.onClick}>
          <div className="row">
            <Errors />
            {this.props.children}
          </div>
        </div>
        <Footer />
      </div>
    );
  }
}

export default Layout;
