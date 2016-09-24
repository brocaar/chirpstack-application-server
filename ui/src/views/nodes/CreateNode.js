import React, { Component } from "react";
import { Link } from "react-router";

import NodeStore from "../../stores/NodeStore";
import NodeForm from "../../components/NodeForm";

class CreateNode extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      node: {},
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(node) {
    NodeStore.createNode(node, (responseData) => {
      this.context.router.push('/');
    }); 
  }

  render() {
    return (
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Nodes</Link></li>
          <li className="active">create node</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeForm node={this.state.node} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateNode;
