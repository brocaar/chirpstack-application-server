import React, { Component } from 'react';

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

  componentDidMount() {
    this.setState({
      configuration: {
        networkServerID: this.props.params.networkServerID,
      },
    });
  }

  onSubmit(configuration) {
    GatewayStore.createChannelConfiguration(configuration, (responseData) => {
      this.context.router.push(`network-servers/${this.props.params.networkServerID}/channel-configurations`);  
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

export default CreateChannelConfiguration;
