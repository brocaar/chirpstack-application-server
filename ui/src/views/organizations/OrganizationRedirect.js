import React, { Component } from "react";

import OrganizationStore from "../../stores/OrganizationStore";
import SessionStore from "../../stores/SessionStore";


class OrganizationRedirect extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  componentDidMount() {
    const organizationID = SessionStore.getOrganizationID();
    if (parseInt(organizationID, 10) === parseInt(organizationID, 10)) {
      this.context.router.push('/organizations/' + organizationID); 
    } else {
      OrganizationStore.getAll("", 1, 0, (totalCount, orgs) => {
        if (orgs.length > 0) {
          this.context.router.push('/organizations/' + orgs[0].id); 
        }    
      });
    }
  }

  render() {
    return(<div></div>);
  }
}

export default OrganizationRedirect;
