import React, { Component } from "react";
import { Link } from "react-router";

import OrganizationSelect from "../../components/OrganizationSelect";
import NodeStore from "../../stores/NodeStore";
import ApplicationStore from "../../stores/ApplicationStore";
import SessionStore from "../../stores/SessionStore";
import NodeActivationForm from "../../components/NodeActivationForm";

class ActivateNode extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      activation: {},
      application: {},
      node: {},
      isApplicationAdmin: false,
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    NodeStore.getNode(this.props.params.applicationID, this.props.params.devEUI, (node) => {
      this.setState({node: node});
    });

    NodeStore.getActivation(this.props.params.applicationID, this.props.params.devEUI, (activation) => {
      this.setState({activation: activation});
    });

    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isApplicationAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isApplicationAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
      });
    });
  }

  onSubmit(activation) {
    NodeStore.activateNode(this.props.params.applicationID, this.props.params.devEUI, activation, (responseData) => {
      this.context.router.push("/organizations/"+this.props.params.organizationID+"/applications/"+this.props.params.applicationID);
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications`}>Applications</Link></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}`}>{this.state.application.name}</Link></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/nodes/${this.props.params.devEUI}/edit`}>{this.state.node.name}</Link></li>
          <li className="active">Activation</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeActivationForm onSubmit={this.onSubmit} node={this.state.node} application={this.state.application} activation={this.state.activation} disabled={!this.state.isApplicationAdmin} />
          </div>
        </div>
      </div>
    );
  }
}

export default ActivateNode;
