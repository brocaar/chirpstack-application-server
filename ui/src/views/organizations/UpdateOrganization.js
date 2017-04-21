import React, { Component } from 'react';
import { Link } from 'react-router';

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
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    OrganizationStore.getOrganization(this.props.params.id, (organization) => {
      this.setState({
        organization: organization,
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
          <li><Link to="/">LoRa Server</Link></li>
          <li><Link to="/organizations">Organizations</Link></li>
          <li>{this.state.organization.name}</li>
          <li className="active">Edit organization</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/organizations/${this.props.params.id}`}><button type="button" className="btn btn-primary">Goto organization</button></Link> &nbsp;
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
