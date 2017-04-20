import React, { Component } from "react";

import OrganizationStore from "../../stores/OrganizationStore";
import SessionStore from "../../stores/SessionStore";


class OrganizationRedirect extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  componentWillMount() {
    const organizationID = SessionStore.getOrganizationID();
    if (organizationID === parseInt(organizationID, 10)) {
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
