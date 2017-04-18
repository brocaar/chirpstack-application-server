import React, { Component } from "react";
import { Link } from "react-router";

import OrganizationSelect from "../../components/OrganizationSelect";
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
      <div>
        <ol className="breadcrumb">
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications`}>Applications</Link></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}`}>{this.state.application.name}</Link></li>
          <li className="active">Create node</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeForm node={this.state.node} application={this.state.application} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateNode;
