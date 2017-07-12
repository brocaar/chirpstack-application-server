import React, { Component } from "react";
import { Link } from 'react-router';

import GatewayStore from "../../stores/GatewayStore";


class ChannelConfigurationLayout extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      configuration: {},
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    GatewayStore.getChannelConfiguration(this.props.params.id, (configuration) => {
      this.setState({
        configuration: configuration,
      });
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this channel-configuration?")) {
      GatewayStore.deleteChannelConfiguration(this.props.params.id, (responseData) => {
        this.context.router.push("/gateways/channelconfigurations");
      });
    }
  }

  render() {
    let activeTab = "";

    if (typeof(this.props.children.props.route.path) !== "undefined") {
      activeTab = this.props.children.props.route.path; 
    }

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/gateways/channelconfigurations">Channel-configurations</Link></li>
          <li className="active">{this.state.configuration.name}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete channel-configuration</button></Link>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={activeTab === "edit" ? "active" : ""}><Link to={`/gateways/channelconfigurations/${this.props.params.id}/edit`}>Channel-configuration</Link></li>
          <li role="presentation" className={activeTab === "edit/extrachannels" ? "active" : ""}><Link to={`/gateways/channelconfigurations/${this.props.params.id}/edit/extrachannels`}>Extra channels</Link></li>
        </ul>
        <hr />
        {this.props.children}
      </div>
    );
  }
}

export default ChannelConfigurationLayout;
