import React, { Component } from "react";

import GatewayStore from "../../stores/GatewayStore";
import ChannelConfigurationForm from "../../components/ChannelConfigurationForm";


class UpdateChannelConfiguration extends Component {
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
    GatewayStore.getChannelConfiguration(this.props.params.id, (configuration) => {
      this.setState({
        configuration: configuration,
      });
    });
  }

  onSubmit(configuration) {
    GatewayStore.updateChannelConfiguration(this.props.params.id, configuration, (responseData) => {
      this.context.router.push("/gateways/channelconfigurations");
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

export default UpdateChannelConfiguration;
