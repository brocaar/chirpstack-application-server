import React, { Component } from 'react';

import GatewayStore from "../../stores/GatewayStore";
import GatewayForm from "../../components/GatewayForm";

class UpdateGateway extends Component {
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
    GatewayStore.getGateway(this.props.params.mac, (gateway) => {
      this.setState({
        gateway: gateway,
      });
    });
  }

  onSubmit(gateway) {
    GatewayStore.updateGateway(this.props.params.mac, gateway, (responseData) => {
      this.context.router.push('/organizations/'+gateway.organizationID+'/gateways/'+gateway.mac);
      window.scrollTo(0, 0);
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <GatewayForm organizationID={this.props.params.organizationID} gateway={this.state.gateway} onSubmit={this.onSubmit} update={true} />
        </div>
      </div>
    );
  }
}

export default UpdateGateway;
