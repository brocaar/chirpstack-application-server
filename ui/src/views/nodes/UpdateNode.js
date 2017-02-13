import React, { Component } from 'react';
import { Link } from 'react-router';

import NodeStore from "../../stores/NodeStore";
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
    };

    this.onSubmit = this.onSubmit.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  componentWillMount() {
    NodeStore.getNode(this.props.params.applicationName, this.props.params.nodeName, (node) => {
      this.setState({node: node});
    });
    ApplicationStore.getApplication(this.props.params.applicationName, (application) => {
      this.setState({application: application});
    });
  }

  onSubmit(node) {
    NodeStore.updateNode(this.props.params.applicationName, this.props.params.nodeName, node, (responseData) => {
      this.context.router.push('/applications/'+this.props.params.applicationName);
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this node?")) {
      NodeStore.deleteNode(this.props.params.applicationName, this.props.params.nodeName, (responseData) => {
        this.context.router.push('/applications/'+this.props.params.applicationName);
      });
    }
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li><Link to={`/applications/${this.props.params.applicationName}`}>{this.state.application.name}</Link></li>
          <li>{this.props.params.nodeName}</li>
          <li className="active">Edit node</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link to={`/applications/${this.props.params.applicationName}/nodes/${this.props.params.nodeName}/activation`} className={(this.state.node.isABP ? '' : 'hidden')}><button type="button" className="btn btn-default">(Re)activate node (ABP)</button></Link> &nbsp;
            <Link to={`/applications/${this.props.params.applicationName}/nodes/${this.props.params.nodeName}/activation`} className={(this.state.node.isABP ? 'hidden' : '')}><button type="button" className="btn btn-default">Show node activation</button></Link> &nbsp;
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete node</button></Link>
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

export default UpdateNode;
