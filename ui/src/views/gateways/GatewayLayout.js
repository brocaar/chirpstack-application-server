import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from 'react-router-dom';

import OrganizationSelect from "../../components/OrganizationSelect";
import GatewayStore from "../../stores/GatewayStore";
import SessionStore from "../../stores/SessionStore";

import GatewayDetails from "./GatewayDetails";
import UpdateGateway from "./UpdateGateway";
import GatewayToken from "./GatewayToken";
import GatewayPing from "./GatewayPing";
import GatewayFrameLogs from "./GatewayFrameLogs";


class GatewayLayout extends Component {
  constructor() {
    super();

    this.state = {
      gateway: {},
      isAdmin: false,
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    GatewayStore.getGateway(this.props.match.params.mac, (gateway) => {
      this.setState({
        gateway: gateway,
      });
    });

    this.setState({
      isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID), 
      });
    });

    GatewayStore.on("change", () => {
      GatewayStore.getGateway(this.props.match.params.mac, (gateway) => {
        this.setState({
          gateway: gateway,
        });
      });
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this gateway?")) {
      GatewayStore.deleteGateway(this.props.match.params.mac, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/gateways`);
      });
    }
  }

  render() {
    let activeTab = this.props.location.pathname.replace(this.props.match.url, '');

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.match.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.match.params.organizationID}/gateways`}>Gateways</Link></li>
          <li className="active">{this.state.gateway.name}</li>
        </ol>
        <div className="clearfix">
          <div className={"btn-group pull-right " + (this.state.isAdmin ? '' : 'hidden')} role="group" aria-label="...">
            <a><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete gateway</button></a>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={activeTab === "" ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.mac}`}>Gateway details</Link></li>
          <li role="presentation" className={(activeTab === "/edit" ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.mac}/edit`}>Gateway configuration</Link></li>
          <li role="presentation" className={(activeTab === "/token" ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.mac}/token`}>Gateway token</Link></li>
          <li role="presentation" className={(activeTab === "/ping" ? 'active' : '') + (this.state.gateway.ping ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.mac}/ping`}>Gateway discovery</Link></li>
          <li role="presentation" className={(activeTab === "/frames" ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.mac}/frames`}>Live LoRaWAN frame logs</Link></li>
        </ul>
        <hr />
        <Switch>
          <Route exact path={this.props.match.path} component={GatewayDetails} />
          <Route exact path={`${this.props.match.path}/edit`} component={UpdateGateway} />
          <Route exact path={`${this.props.match.path}/token`} component={GatewayToken} />
          <Route exact path={`${this.props.match.path}/ping`} component={GatewayPing} />
          <Route exact path={`${this.props.match.path}/frames`} component={GatewayFrameLogs} />
        </Switch>
      </div>
    );
  }
}

export default withRouter(GatewayLayout);
