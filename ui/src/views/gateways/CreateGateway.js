import React, { Component } from "react";

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
        <div className="panel panel-default">
        <div className="panel-heading">
          <h3 className="panel-title">Create gateway</h3>
        </div>
        <div className="panel-body">
          <GatewayForm organizationID={this.props.params.organizationID} gateway={this.state.gateway} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default CreateGateway;
