import React, { Component } from 'react';

import SessionStore from "../stores/SessionStore";

class Footer extends Component {
  constructor() {
    super();
    this.state = {
      footer: null,
    }
  }

  componentDidMount() {
    this.setState({
      footer: SessionStore.getFooter(),
    });

    SessionStore.on("change", () => {
      this.setState({
        footer: SessionStore.getFooter(),
      });
    });
  }

  render() {
    return (
      <footer className="footer">
        <div className="container">
          <p className="text-muted" dangerouslySetInnerHTML={{__html: this.state.footer}} />
        </div>
      </footer>
    );
  }
}

export default Footer;