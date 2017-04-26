import React, { Component } from 'react';
import { Link } from 'react-router';

import OrganizationStore from "../../stores/OrganizationStore";
import SessionStore from "../../stores/SessionStore";
import OrganizationForm from "../../components/OrganizationForm";

class UpdateOrganization extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      organization: {},
      isAdmin: false,
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    OrganizationStore.getOrganization(this.props.params.id, (organization) => {
      this.setState({
        organization: organization,
      });
    });

    this.setState({
      isAdmin: SessionStore.isAdmin(),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: SessionStore.isAdmin(),
      });
    });
  }

  onSubmit(organization) {
    OrganizationStore.updateOrganization(this.props.params.id, organization, (responseData) => {
      this.context.router.push('organizations');
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this organization?")) {
      OrganizationStore.deleteOrganization(this.props.params.id, (responseData) => {
        this.context.router.push('/organizations');
      });
    }
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><Link to={`/organizations/${this.props.params.id}`}>{this.state.organization.displayName}</Link></li>
          <li className="active">Edit organization</li>
        </ol>
        <div className="clearfix">
          <div className={'btn-group pull-right ' + (this.state.isAdmin ? '' : 'hidden')} role="group" aria-label="...">
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete organization</button></Link>
          </div>
        </div>
        <hr />
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

export default UpdateOrganization;
