import React, { Component } from "react";

import DeviceProfileStore from "../../stores/DeviceProfileStore";
import NodeStore from "../../stores/NodeStore";
import NodeForm from "../../components/NodeForm";


class CreateNode extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

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
        applicationID: this.props.params.applicationID,
      },
    });
  }

  onSubmit(node) {
    NodeStore.createNode(this.props.params.applicationID, node, (responseData) => {
      DeviceProfileStore.getDeviceProfile(node.deviceProfileID, (deviceProfile) => {
        if (deviceProfile.deviceProfile.supportsJoin) {
          this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+this.props.params.applicationID+"/nodes/"+node.devEUI+"/keys");
        } else {
          this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+this.props.params.applicationID+"/nodes/"+node.devEUI+"/activate");
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
          <NodeForm node={this.state.node} applicationID={this.props.params.applicationID} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default CreateNode;
