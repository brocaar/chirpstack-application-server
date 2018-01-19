import React, { Component } from "react";
import { Link, withRouter } from "react-router-dom";

import OrganizationStore from "../../stores/OrganizationStore";
import OrganizationForm from "../../components/OrganizationForm";


class CreateOrganization extends Component {
  constructor() {
    super();

    this.state = {
      organization: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(organization) {
	  OrganizationStore.createOrganization(organization, (responseData) => {
      this.props.history.push("/organizations");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li className="active">Create organization</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <OrganizationForm organization={this.state.organization} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(CreateOrganization);
