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

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Nodes</Link></li>
          <li className="active">{this.props.params.devEUI}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to="/nodes/create"><button type="button" className="btn btn-default">Create node</button></Link>
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
