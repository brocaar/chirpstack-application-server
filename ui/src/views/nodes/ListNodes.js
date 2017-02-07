import React, { Component } from 'react';
import { Link } from 'react-router';

import NodeStore from "../../stores/NodeStore";
import ApplicationStore from "../../stores/ApplicationStore";

class NodeRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/applications/${this.props.application.name}/nodes/${this.props.node.devEUI}/edit`}>{this.props.node.devEUI}</Link></td>
        <td>{this.props.node.name}</td>
      </tr>
    );
  }
}

class ListNodes extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      application: {},
      nodes: [],
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    NodeStore.getAll(this.props.params.applicationName, (nodes) => {
      this.setState({nodes: nodes});
    });

    ApplicationStore.getApplication(this.props.params.applicationName, (application) => {
      this.setState({application: application});
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this application?")) {
      ApplicationStore.deleteApplication(this.props.params.applicationName, (responseData) => {
        this.context.router.push("/applications");
      }); 
    }
  }

  render() {
    const NodeRows = this.state.nodes.map((node, i) => <NodeRow key={node.devEUI} node={node} application={this.state.application} />);

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li>{this.state.application.name}</li>
          <li className="active">Nodes</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/applications/${this.props.params.applicationName}/nodes/create`}><button type="button" className="btn btn-default">Create node</button></Link> &nbsp;
            <Link to={`/applications/${this.props.params.applicationName}/edit`}><button type="button" className="btn btn-default">Edit application</button></Link> &nbsp;
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete application</button></Link>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>Device EUI</th>
                  <th>Device name</th>
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
