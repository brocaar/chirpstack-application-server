import React, { Component } from "react";

import NodeStore from "../../stores/NodeStore";
import NodeForm from "../../components/NodeForm";
import ApplicationStore from "../../stores/ApplicationStore";

class CreateNode extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      application: {},
      node: {
        useApplicationSettings: true,
      },
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({application: application});
    });
  }

  onSubmit(node) {
    node.applicationID = this.props.params.applicationID;
    NodeStore.createNode(this.props.params.applicationID, node, (responseData) => {
      this.context.router.push('/organizations/'+this.props.params.organizationID+'/applications/'+this.props.params.applicationID);
    }); 
  }

  render() {
    return (
      <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title">Create node</h3>
        </div>
        <div className="panel-body">
          <NodeForm node={this.state.node} application={this.state.application} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default CreateNode;
