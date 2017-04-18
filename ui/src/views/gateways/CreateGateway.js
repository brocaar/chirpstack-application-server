import React, { Component } from "react";
import { Link } from "react-router";

import OrganizationSelect from "../../components/OrganizationSelect";
import GatewayStore from "../../stores/GatewayStore";
import GatewayForm from "../../components/GatewayForm";

class CreateGateway extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      gateway: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentWillMount() {
    this.setState({
      gateway: {organizationID: this.props.params.organizationID},
    });
  }

  onSubmit(gateway) {
    GatewayStore.createGateway(gateway, (responseData) => {
      this.context.router.push("/organizations/"+this.props.params.organizationID+"/gateways");
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/gateways`}>Gateways</Link></li>
          <li className="active">Create gateway</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <GatewayForm gateway={this.state.gateway} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateGateway;
