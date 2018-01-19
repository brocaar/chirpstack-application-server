import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import OrganizationStore from "../../stores/OrganizationStore";
import OrganizationForm from "../../components/OrganizationForm";


class UpdateOrganization extends Component {
  constructor() {
    super();

    this.state = {
      organization: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    OrganizationStore.getOrganization(this.props.match.params.organizationID, (organization) => {
      this.setState({
        organization: organization,
      });
    });
  }

  onSubmit(organization) {
    OrganizationStore.updateOrganization(this.props.match.params.organizationID, organization, (responseData) => {
      this.props.history.push('/organizations');
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <OrganizationForm organization={this.state.organization} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateOrganization);
