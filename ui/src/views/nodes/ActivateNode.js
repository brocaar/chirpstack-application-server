import React, { Component } from "react";
import { Link } from "react-router";

import NodeStore from "../../stores/NodeStore";
import NodeActivationForm from "../../components/NodeActivationForm";

class ActivateNode extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      activation: {},
      application: {},
      node: {},
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    NodeStore.getNode(this.props.params.applicationName, this.props.params.nodeName, (node) => {
      this.setState({node: node});
    });

    NodeStore.getActivation(this.props.params.applicationName, this.props.params.nodeName, (activation) => {
      this.setState({activation: activation});
    });
  }

  onSubmit(activation) {
    NodeStore.activateNode(this.props.params.applicationName, this.props.params.nodeName, activation, (responseData) => {
      this.context.router.push("/applications/"+this.props.params.applicationName);
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li><Link to={`/applications/${this.props.params.applicationName}`}>{this.props.params.applicationName}</Link></li>
          <li><Link to={`/applications/${this.props.params.applicationName}/nodes/${this.props.params.nodeName}/edit`}>{this.props.params.nodeName}</Link></li>
          <li className="active">Activation</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeActivationForm onSubmit={this.onSubmit} node={this.state.node} application={this.state.application} activation={this.state.activation} />
          </div>
        </div>
      </div>
    );
  }
}

export default ActivateNode;
