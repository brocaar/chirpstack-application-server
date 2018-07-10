import { Component } from "react";

import SessionStore from "../stores/SessionStore";


class Admin extends Component {
  constructor() {
    super();
    this.state = {
      admin: false,
    };

    this.setIsAdmin = this.setIsAdmin.bind(this);
  }

  componentDidMount() {
    SessionStore.on("change", this.setIsAdmin);
    this.setIsAdmin();
  }

  componentWillUnmount() {
    SessionStore.removeListener("change", this.setIsAdmin);
  }

  componentDidUpdate(prevProps) {
    if (prevProps === this.props) {
      return;
    }

    this.setIsAdmin();
  }

  setIsAdmin() {
    if (this.props.organizationID !== undefined) {
      this.setState({
        admin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.organizationID),
      });
    } else {
      this.setState({
        admin: SessionStore.isAdmin(),
      });
    }
  }

  render() {
    if (this.state.admin) {
      return(this.props.children);
    }

    return(null);
  }
}

export default Admin;
