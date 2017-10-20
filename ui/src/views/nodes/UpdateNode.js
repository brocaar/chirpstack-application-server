import React, { Component } from 'react';

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

  render() {
    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeForm applicationID={this.props.params.applicationID} node={this.state.node} onSubmit={this.onSubmit} disabled={!this.state.isAdmin} application={this.state.application} />
          </div>
        </div>
      </div>
    );
  }
}

export default UpdateNode;
