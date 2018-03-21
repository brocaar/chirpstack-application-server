import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from 'react-router-dom';

import OrganizationSelect from "../../components/OrganizationSelect";
import NodeStore from "../../stores/NodeStore";
import ApplicationStore from "../../stores/ApplicationStore";
import DeviceProfileStore from "../../stores/DeviceProfileStore";
import SessionStore from "../../stores/SessionStore";

import UpdateNode from './UpdateNode';
import ActivateNode from "./ActivateNode";
import NodeEventLogs from "./NodeEventLogs";
import NodeFrameLogs from "./NodeFrameLogs";
import NodeKeys from "./NodeKeys";
import NodeActivation from "./NodeActivation";


class NodeLayout extends Component {
  constructor() {
    super();

    this.state = {
      application: {},
      node: {},
      deviceProfile: {
        deviceProfile: {},
      },
      isAdmin: false,
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    NodeStore.getNode(this.props.match.params.applicationID, this.props.match.params.devEUI, (node) => {
      this.setState({node: node});

      DeviceProfileStore.getDeviceProfile(this.state.node.deviceProfileID, (deviceProfile) => {
        this.setState({
          deviceProfile: deviceProfile,
        });
      });
    });

    ApplicationStore.getApplication(this.props.match.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
      });
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this node?")) {
      NodeStore.deleteNode(this.props.match.params.applicationID, this.props.match.params.devEUI, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
      });
    }
  }

  render() {
    let activeTab = this.props.location.pathname.replace(this.props.match.url, '');

    return (
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.match.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.match.params.organizationID}/applications`}>Applications</Link></li>
          <li><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`}>{this.state.application.name}</Link></li>
          <li className="active">{this.state.node.name}</li>
        </ol>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <a><button type="button" className={"btn btn-danger " + (this.state.isAdmin ? '' : 'hidden')} onClick={this.onDelete}>Delete device</button></a>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={activeTab === "/edit" ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${this.props.match.params.devEUI}/edit`}>Device configuration</Link></li>
          <li role="presentation" className={(activeTab === "/keys" ? 'active' : '') + (this.state.deviceProfile.deviceProfile.supportsJoin && this.state.isAdmin ? "" : "hidden")}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${this.props.match.params.devEUI}/keys`}>Device keys (OTAA)</Link></li>
          <li role="presentation" className={(activeTab === "/activate" ? 'active' : '') + (!this.state.deviceProfile.deviceProfile.supportsJoin && this.state.isAdmin ? "": "hidden")}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${this.props.match.params.devEUI}/activate`}>Activate device (ABP)</Link></li>
          <li role="presentation" className={activeTab === "/activation" ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${this.props.match.params.devEUI}/activation`}>Device activation</Link></li>
          <li role="presentation" className={activeTab === "/events" ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${this.props.match.params.devEUI}/events`}>Live event logs</Link></li>
          <li role="presentation" className={activeTab === "/frames" ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${this.props.match.params.devEUI}/frames`}>Live LoRaWAN frame logs</Link></li>
        </ul>
        <hr />
        <Switch>
          <Route exact path={`${this.props.match.path}/edit`} component={UpdateNode} />
          <Route exact path={`${this.props.match.path}/activate`} component={ActivateNode} />
          <Route exact path={`${this.props.match.path}/frames`} component={NodeFrameLogs} />
          <Route exact path={`${this.props.match.path}/events`} component={NodeEventLogs} />
          <Route exact path={`${this.props.match.path}/keys`} component={NodeKeys} />
          <Route exact path={`${this.props.match.path}/activation`} component={NodeActivation} />
        </Switch>
      </div>
    );
  }
}

export default withRouter(NodeLayout);
