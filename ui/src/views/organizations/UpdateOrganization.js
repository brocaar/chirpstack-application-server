import React, { Component } from 'react';

import OrganizationStore from "../../stores/OrganizationStore";
import OrganizationForm from "../../components/OrganizationForm";


class UpdateOrganization extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      organization: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    OrganizationStore.getOrganization(this.props.params.organizationID, (organization) => {
      this.setState({
        organization: organization,
      });
    });
  }

  onSubmit(organization) {
    OrganizationStore.updateOrganization(this.props.params.organizationID, organization, (responseData) => {
      this.context.router.push('organizations');
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

export default UpdateOrganization;
