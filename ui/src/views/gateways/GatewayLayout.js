import React, { Component } from "react";
import { Link } from 'react-router';

import OrganizationSelect from "../../components/OrganizationSelect";
import GatewayStore from "../../stores/GatewayStore";
import SessionStore from "../../stores/SessionStore";

class GatewayLayout extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      gateway: {},
      isAdmin: false,
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    GatewayStore.getGateway(this.props.params.mac, (gateway) => {
      this.setState({
        gateway: gateway,
      });
    });

    this.setState({
      isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID), 
      });
    });

    GatewayStore.on("change", () => {
      GatewayStore.getGateway(this.props.params.mac, (gateway) => {
        this.setState({
          gateway: gateway,
        });
      });
    });
  }

  onDelete() {
    if (confirm("Are you sure you want to delete this gateway?")) {
      GatewayStore.deleteGateway(this.props.params.mac, (responseData) => {
        this.context.router.push("/organizations/"+this.props.params.organizationID+"/gateways");
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
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.params.organizationID}/gateways`}>Gateways</Link></li>
          <li className="active">{this.state.gateway.name}</li>
        </ol>
        <div className="clearfix">
          <div className={"btn-group pull-right " + (this.state.isAdmin ? '' : 'hidden')} role="group" aria-label="...">
            <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete gateway</button></Link>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={activeTab === "" ? 'active' : ''}><Link to={`/organizations/${this.props.params.organizationID}/gateways/${this.props.params.mac}`}>Gateway details</Link></li>
          <li role="presentation" className={(activeTab === "edit" ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/gateways/${this.props.params.mac}/edit`}>Gateway configuration</Link></li>
          <li role="presentation" className={(activeTab === "token" ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/gateways/${this.props.params.mac}/token`}>Gateway token</Link></li>
          <li role="presentation" className={(activeTab === "ping" ? 'active' : '') + (this.state.gateway.ping ? '' : 'hidden')}><Link to={`/organizations/${this.props.params.organizationID}/gateways/${this.props.params.mac}/ping`}>Gateway discovery</Link></li>
        </ul>
        <hr />
        {this.props.children}
      </div>
    );
  }
}

export default GatewayLayout;
