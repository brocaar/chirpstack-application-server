import React, { Component } from 'react';
import { Link } from 'react-router';

import NodeStore from "../../stores/NodeStore";

class NodeRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/nodes/${this.props.node.devEUI}`}>{this.props.node.devEUI}</Link></td>
        <td>{this.props.node.name}</td>
        <td>{this.props.node.appEUI}</td>
      </tr>
    );
  }
}

class ListNodes extends Component {
  constructor() {
    super();
    this.state = {
      nodes: [],
    };
    NodeStore.getAll((nodes) => {
      this.setState({nodes: nodes});
    });
  }

  render() {
    const NodeRows = this.state.nodes.map((node, i) => <NodeRow key={node.devEUI} node={node} />);

    return(
      <div>
        <ol className="breadcrumb">
          <li className="active">Nodes</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to="/nodes/create"><button type="button" className="btn btn-default">Create node</button></Link> &nbsp;
            <Link to="/channels"><button type="button" className="btn btn-default">Channel lists</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>DevEUI</th>
                  <th>Name</th>
                  <th>AppEUI</th>
                </tr>
              </thead>
              <tbody>
                {NodeRows}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}

export default ListNodes;
