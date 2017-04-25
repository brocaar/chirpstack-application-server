import React, { Component } from 'react';
import { Link } from 'react-router';

import OrganizationSelect from "../../components/OrganizationSelect";
import NodeStore from "../../stores/NodeStore";
import SessionStore from "../../stores/SessionStore";
import NodeForm from "../../components/NodeForm";
import ApplicationStore from "../../stores/ApplicationStore";

class UpdateNode extends Component {
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

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
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

  onSubmit(node) {
    node.applicationID = this.props.params.applicationID;
    NodeStore.updateNode(this.props.params.applicationID, this.props.params.devEUI, node, (responseData) => {
      this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+this.props.params.applicationID);
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
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications`}>Applications</Link></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}`}>{this.state.application.name}</Link></li>
          <li>{this.state.node.name}</li>
          <li className="active">Edit node</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/nodes/${this.props.params.devEUI}/activation`} className={(this.state.node.isABP ? '' : 'hidden')}><button type="button" className="btn btn-default">ABP activation</button></Link> &nbsp;
            <Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/nodes/${this.props.params.devEUI}/activation`} className={(this.state.node.isABP ? 'hidden' : '')}><button type="button" className="btn btn-default">Show node activation</button></Link> &nbsp;
            <Link><button type="button" className={"btn btn-danger " + (this.state.isAdmin ? '' : 'hidden')} onClick={this.onDelete}>Delete node</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeForm node={this.state.node} onSubmit={this.onSubmit} disabled={!this.state.isAdmin} application={this.state.application} />
          </div>
        </div>
      </div>
    );
  }
}

export default UpdateNode;
