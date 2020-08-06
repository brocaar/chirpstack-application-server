import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import OrganizationStore from "../../stores/OrganizationStore";
import SessionStore from "../../stores/SessionStore";


class OrganizationRedirect extends Component {
  componentDidMount() {
    const organizationID = SessionStore.getOrganizationID();
    if (organizationID !== undefined && organizationID !== null && organizationID !== "") {
      this.props.history.push(`/organizations/${organizationID}`);
    } else {
      OrganizationStore.list("", 1, 0, resp => {
        if (resp.result.length > 0) {
          this.props.history.push(`/organizations/${resp.result[0].id}`);
        }
      });
    }
  }

  render() {
    return(<div></div>);
  }
}

export default withRouter(OrganizationRedirect);
