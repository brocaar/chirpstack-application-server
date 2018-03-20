import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import NodeStore from "../../stores/NodeStore";
import SessionStore from "../../stores/SessionStore";
import NodeForm from "../../components/NodeForm";
import ApplicationStore from "../../stores/ApplicationStore";


class UpdateNode extends Component {
  constructor() {
    super();

    this.state = {
      application: {},
      node: {},
      isAdmin: false,
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    NodeStore.getNode(this.props.match.params.applicationID, this.props.match.params.devEUI, (node) => {
      this.setState({node: node});
    });
    ApplicationStore.getApplication(this.props.match.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
      });
    });
  }

  onSubmit(node) {
    node.applicationID = this.props.match.params.applicationID;
    NodeStore.updateNode(this.props.match.params.applicationID, this.props.match.params.devEUI, node, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
    });
  }

  render() {
    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeForm organizationID={this.props.match.params.organizationID} applicationID={this.props.match.params.applicationID} node={this.state.node} onSubmit={this.onSubmit} disabled={!this.state.isAdmin} application={this.state.application} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateNode);
