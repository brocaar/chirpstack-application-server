import React, { Component } from 'react';
import { Link } from 'react-router';

import GatewayStore from "../../stores/GatewayStore";
import ChannelConfigurationForm from "../../components/ChannelConfigurationForm";


class CreateChannelConfiguration extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      configuration: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(configuration) {
    GatewayStore.createChannelConfiguration(configuration, (responseData) => {
      this.context.router.push("/gateways/channelconfigurations");  
    });
  }

  render() {
    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/gateways/channelconfigurations">Channel-configurations</Link></li>
          <li className="active">Create</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <ChannelConfigurationForm configuration={this.state.configuration} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default CreateChannelConfiguration;
