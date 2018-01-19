import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import GatewayStore from "../../stores/GatewayStore";
import ChannelConfigurationForm from "../../components/ChannelConfigurationForm";


class CreateChannelConfiguration extends Component {
  constructor() {
    super();

    this.state = {
      configuration: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    this.setState({
      configuration: {
        networkServerID: this.props.match.params.networkServerID,
      },
    });
  }

  onSubmit(configuration) {
    GatewayStore.createChannelConfiguration(configuration, (responseData) => {
      this.props.history.push(`/network-servers/${this.props.match.params.networkServerID}/channel-configurations`);
    });
  }

  render() {
    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-heading">
            Create channel-configuration
          </div>
          <div className="panel-body">
            <ChannelConfigurationForm configuration={this.state.configuration} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(CreateChannelConfiguration);
