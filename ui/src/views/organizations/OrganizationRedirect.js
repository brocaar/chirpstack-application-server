import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import OrganizationStore from "../../stores/OrganizationStore";
import SessionStore from "../../stores/SessionStore";


class OrganizationRedirect extends Component {
  componentDidMount() {
    const organizationID = SessionStore.getOrganizationID();
    if (!isNaN(parseInt(organizationID, 10))) {
      this.props.history.push('/organizations/' + organizationID); 
    } else {
      OrganizationStore.getAll("", 1, 0, (totalCount, orgs) => {
        if (orgs.length > 0) {
          this.props.history.push('/organizations/' + orgs[0].id); 
        }    
      });
    }
  }

  render() {
    return(<div></div>);
  }
}

export default withRouter(OrganizationRedirect);
