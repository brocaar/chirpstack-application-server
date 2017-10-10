import React, { Component } from "react";
import { Link } from 'react-router';

import SessionStore from "../../stores/SessionStore";
import OrganizationSelect from "../../components/OrganizationSelect";
import OrganizationStore from "../../stores/OrganizationStore";

class OrganizationLayout extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      isAdmin: false,
      isGlobalAdmin: false,
      organization: {},
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
      isGlobalAdmin: SessionStore.isAdmin(),
    });

    OrganizationStore.getOrganization(this.props.params.organizationID, (org) => {
      this.setState({
        organization: org,
      });
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
        isGlobalAdmin: SessionStore.isAdmin(),
      });
    });
  }

  componentWillReceiveProps(props) {
    OrganizationStore.getOrganization(props.params.organizationID, (org) => {
      this.setState({
        organization: org,
      });
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this organization?")) {
      OrganizationStore.deleteOrganization(this.props.params.organizationID, (responseData) => {
        this.context.router.push('/organizations');
      });
    }
  }


  render() {
    let activeTab = "";

    if (typeof(this.props.children.props.route.path) !== "undefined") {
      activeTab = this.props.children.props.route.path; 
    }

    return (
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
        </ol>
        <div className={(this.state.isGlobalAdmin ? '' : 'hidden')}>
          <div className="clearfix">
            <div className="btn-group pull-right" role="group" aria-label="...">
              <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete organization</button></Link>
            </div>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={(activeTab === "" || activeTab.startsWith("applications")) ? 'active' : ''}><Link to={`/organizations/${this.props.params.organizationID}`}>Applications</Link></li>
          <li role="presentation" className={(activeTab.startsWith("gateways") ? 'active' : '') + (this.state.organization.canHaveGateways ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/gateways`}>Gateways</Link></li>
          <li role="presentation" className={(activeTab === "edit" ? 'active': '') + (this.state.isGlobalAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/edit`}>Organization configuration</Link></li>
          <li role="presentation" className={(activeTab.startsWith("users") ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/users`}>Organization users</Link></li>
          <li role="presentation" className={activeTab.startsWith("service-profiles") ? 'active' : ''}><Link to={`/organizations/${this.props.params.organizationID}/service-profiles`}>Service profiles</Link></li>
        </ul>
        <hr />
        {this.props.children} 
      </div>
    );
  }
}

export default OrganizationLayout;
