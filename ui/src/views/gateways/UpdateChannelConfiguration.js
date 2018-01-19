import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import GatewayStore from "../../stores/GatewayStore";
import ChannelConfigurationForm from "../../components/ChannelConfigurationForm";


class UpdateChannelConfiguration extends Component {
  constructor() {
    super();

    this.state = {
      configuration: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    GatewayStore.getChannelConfiguration(this.props.match.params.networkServerID, this.props.match.params.channelConfigurationID, (configuration) => {
      this.setState({
        configuration: configuration,
      });
    });
  }

  onSubmit(configuration) {
    GatewayStore.updateChannelConfiguration(this.props.match.params.channelConfigurationID, configuration, (responseData) => {
      this.props.history.push(`/network-servers/${this.props.match.params.networkServerID}/channel-configurations`);
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <ChannelConfigurationForm configuration={this.state.configuration} onSubmit={this.onSubmit} />
        </div>
      </div>
    );
  }
}

export default withRouter(UpdateChannelConfiguration);
