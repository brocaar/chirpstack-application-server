import React, { Component } from "react";
import { Link } from 'react-router';

import OrganizationSelect from "../../components/OrganizationSelect";
import NodeStore from "../../stores/NodeStore";
import ApplicationStore from "../../stores/ApplicationStore";
import SessionStore from "../../stores/SessionStore";

class NodeLayout extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      application: {},
      node: {},
      isAdmin: false,
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    NodeStore.getNode(this.props.params.applicationID, this.props.params.devEUI, (node) => {
      this.setState({node: node});
    });
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
    if (confirm("Are you sure you want to delete this node?")) {
      NodeStore.deleteNode(this.props.params.applicationID, this.props.params.devEUI, (responseData) => {
        this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+this.props.params.applicationID);
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
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}`}>{this.state.application.name}</Link></li>
          <li className="active">{this.state.node.name}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link><button type="button" className={"btn btn-danger " + (this.state.isAdmin ? '' : 'hidden')} onClick={this.onDelete}>Delete node</button></Link>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={activeTab === "edit" ? 'active' : ''}><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/nodes/${this.props.params.devEUI}/edit`}>Node settings</Link></li>
          <li role="presentation" className={activeTab === "activation" ? 'active' : ''}><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/nodes/${this.props.params.devEUI}/activation`}>Node activation</Link></li>
          <li role="presentation" className={activeTab === "frames" ? 'active' : ''}><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/nodes/${this.props.params.devEUI}/frames`}>Frame logs</Link></li>
        </ul>
        <hr />
        {this.props.children}
      </div>
    );
  }
}

export default NodeLayout;
