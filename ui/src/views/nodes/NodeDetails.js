import React, { Component } from 'react';
import { Link } from 'react-router';

import NodeStore from "../../stores/NodeStore";
import NodeForm from "../../components/NodeForm";

class NodeDetails extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      node: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    NodeStore.getNode(this.props.params.devEUI, (node) => {
      this.setState({node: node});
    });
  }

  onSubmit(node) {
    NodeStore.updateNode(this.props.params.devEUI, node, (responseData) => {
      this.context.router.push('/');
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this node?")) {
      NodeStore.deleteNode(this.props.params.devEUI, (responseData) => {
        this.context.router.push("/");
      });
    } 
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li>Nodes</li>
          <li className="active">{this.props.params.devEUI}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete</button>
          </div>
        </div>
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

export default NodeDetails;
