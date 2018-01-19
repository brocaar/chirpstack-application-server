import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import GatewayStore from "../../stores/GatewayStore";


class ChannelConfigurationRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/network-servers/${this.props.configuration.networkServerID}/channel-configurations/${this.props.configuration.id}/edit`}>{this.props.configuration.name}</Link></td>
      </tr>
    );
  }
}

class ListChannelConfigurations extends Component {
  constructor() {
    super();

    this.state = {
      channelConfigurations: [],
    };
  }

  componentDidMount() {
    GatewayStore.getAllChannelConfigurations(this.props.match.params.networkServerID, (channelConfigurations) => {
      this.setState({
        channelConfigurations: channelConfigurations,
      });
    });  
  }

  render() {
    const ConfigRows = this.state.channelConfigurations.map((conf, i) => <ChannelConfigurationRow key={conf.id} configuration={conf} />);

    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-heading clearfix">
            <div className="btn-group pull-right">
              <Link to={`/network-servers/${this.props.match.params.networkServerID}/channel-configurations/create`}><button type="button" className="btn btn-default btn-sm">Create channel-configuration</button></Link>
            </div>
          </div>
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>Name</th>
                </tr>
              </thead>
              <tbody>
                {ConfigRows}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}

export default ListChannelConfigurations;
