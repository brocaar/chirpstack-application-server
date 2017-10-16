import React, { Component } from 'react';

import dispatcher from "../dispatcher";
import SessionStore from "../stores/SessionStore";

class Footer extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      registration: SessionStore.getRegistration(),
      footer: SessionStore.getFooter(),
    }
  }

  componentWillMount() {
    SessionStore.on("change", () => {
      this.setState({
        registration: SessionStore.getRegistration(),
        footer: SessionStore.getFooter(),
      });
    });
  }

  render() {
    return (
      <div className="footer">
          <div className="footerRegistration" dangerouslySetInnerHTML={{ __html: ( typeof(this.state.registration) === "undefined" ? "" : this.state.registration) }} />
          <div className="footerRight" dangerouslySetInnerHTML={{ __html: ( typeof(this.state.registration) === "undefined" ? "" : this.state.footer) }} />
      </div>
    );
  }
}

export default Footer;
