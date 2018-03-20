import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import DeviceProfileStore from "../../stores/DeviceProfileStore";
import NodeStore from "../../stores/NodeStore";
import NodeForm from "../../components/NodeForm";


class CreateNode extends Component {
  constructor() {
    super();
    this.state = {
      application: {},
      node: {},
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    this.setState({
      node: {
        applicationID: this.props.match.params.applicationID,
      },
    });
  }

  onSubmit(node) {
    NodeStore.createNode(this.props.match.params.applicationID, node, (responseData) => {
      DeviceProfileStore.getDeviceProfile(node.deviceProfileID, (deviceProfile) => {
        if (deviceProfile.deviceProfile.supportsJoin) {
          this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${node.devEUI}/keys`);
        } else {
          this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${node.devEUI}/activate`);
        }
      });
    }); 
  }

  render() {
    return (
      <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title">Create device</h3>
        </div>
        <div className="panel-body">
          <NodeForm node={this.state.node} organizationID={this.props.match.params.organizationID} applicationID={this.props.match.params.applicationID} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(CreateNode);
