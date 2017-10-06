import React, { Component } from "react";
import { Link } from 'react-router';

import SessionStore from "../../stores/SessionStore";
import ApplicationStore from "../../stores/ApplicationStore";
import OrganizationSelect from "../../components/OrganizationSelect";


class ApplicationLayout extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {  
    super();

    this.state = {
      application: {},
      isAdmin: false,
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
      });
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this application?")) {
      ApplicationStore.deleteApplication(this.props.params.applicationID, (responseData) => {
        this.context.router.push("/organizations/"+this.props.params.organizationID+"/applications");
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
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications`}>Applications</Link></li>
          <li className="active">{this.state.application.name}</li>
        </ol>
        <div className={(this.state.isAdmin ? '' : 'hidden')}>
          <div className="clearfix">
            <div className="btn-group pull-right" role="group" aria-label="...">
              <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete application</button></Link>
            </div>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={(activeTab === "" || activeTab === "nodes/create") ? 'active' : ''}><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}`}>Nodes</Link></li>
          <li role="presentation" className={(activeTab === "edit" ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/edit`}>Application configuration</Link></li>
          <li role="presentation" className={((activeTab === "users" || activeTab === "users/:userID/edit" || activeTab === "users/create") ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/users`}>Application users</Link></li>
          <li role="presentation" className={((activeTab === "integrations" || activeTab === "integrations/create" || activeTab === "integrations/http") ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/integrations`}>Integrations</Link></li>
        </ul>
        <hr />
        {this.props.children}
      </div>
    );
  }
}

export default ApplicationLayout;
