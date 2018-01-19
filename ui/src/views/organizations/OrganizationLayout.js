import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from 'react-router-dom';

import SessionStore from "../../stores/SessionStore";
import OrganizationSelect from "../../components/OrganizationSelect";
import OrganizationStore from "../../stores/OrganizationStore";

// applications
import ListApplications from '../applications/ListApplications';
import CreateApplication from "../applications/CreateApplication";

// organizations
import UpdateOrganization from './UpdateOrganization';
import OrganizationUsers from './OrganizationUsers';
import CreateOrganizationUser from './CreateOrganizationUser';
import UpdateOrganizationUser from './UpdateOrganizationUser';

// gateways
import ListGateways from "../gateways/ListGateways";
import CreateGateway from "../gateways/CreateGateway";

// service-profiles
import ListServiceProfiles from "../serviceprofiles/ListServiceProfiles";
import CreateServiceProfile from "../serviceprofiles/CreateServiceProfile";
import UpdateServiceProfile from "../serviceprofiles/UpdateServiceProfile";

// device-profiles
import ListDeviceProfiles from "../deviceprofiles/ListDeviceProfiles";
import CreateDeviceProfile from "../deviceprofiles/CreateDeviceProfile";
import UpdateDeviceProfile from "../deviceprofiles/UpdateDeviceProfile";


class OrganizationLayout extends Component {
  constructor() {
    super();

    this.state = {
      isAdmin: false,
      isGlobalAdmin: false,
      organization: {},
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
      isGlobalAdmin: SessionStore.isAdmin(),
    });

    OrganizationStore.getOrganization(this.props.match.params.organizationID, (org) => {
      this.setState({
        organization: org,
      });
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
        isGlobalAdmin: SessionStore.isAdmin(),
      });
    });
  }

  componentWillReceiveProps(props) {
    OrganizationStore.getOrganization(props.match.params.organizationID, (org) => {
      this.setState({
        organization: org,
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(props.match.params.organizationID)),
        isGlobalAdmin: SessionStore.isAdmin(),
      });
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this organization?")) {
      OrganizationStore.deleteOrganization(this.props.match.params.organizationID, (responseData) => {
        this.props.history.push('/organizations');
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
        </ol>
        <div className={(this.state.isGlobalAdmin ? '' : 'hidden')}>
          <div className="clearfix">
            <div className="btn-group pull-right" role="group" aria-label="...">
              <a><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete organization</button></a>
            </div>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={(activeTab === "" || activeTab.startsWith("/applications")) ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}`}>Applications</Link></li>
          <li role="presentation" className={(activeTab.startsWith("/gateways") ? 'active' : '') + (this.state.organization.canHaveGateways ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/gateways`}>Gateways</Link></li>
          <li role="presentation" className={(activeTab === "/edit" ? 'active': '') + (this.state.isGlobalAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/edit`}>Organization configuration</Link></li>
          <li role="presentation" className={(activeTab.startsWith("/users") ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/users`}>Organization users</Link></li>
          <li role="presentation" className={activeTab.startsWith("/service-profiles") ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/service-profiles`}>Service profiles</Link></li>
          <li role="presentation" className={activeTab.startsWith("/device-profiles") ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/device-profiles`}>Device profiles</Link></li>
        </ul>
        <hr />
        <Switch>
          <Route exact path={this.props.match.path} component={ListApplications} />
          <Route exact path={`${this.props.match.path}/edit`} component={UpdateOrganization} />
          <Route exact path={`${this.props.match.path}/applications`} component={ListApplications} />
          <Route exact path={`${this.props.match.path}/applications/create`} component={CreateApplication} />
          <Route exact path={`${this.props.match.path}/gateways`} component={ListGateways} />
          <Route exact path={`${this.props.match.path}/gateways/create`} component={CreateGateway} />
          <Route exact path={`${this.props.match.path}/users`} component={OrganizationUsers} />
          <Route exact path={`${this.props.match.path}/users/create`} component={CreateOrganizationUser} />
          <Route exact path={`${this.props.match.path}/users/:userID/edit`} component={UpdateOrganizationUser} />
          <Route exact path={`${this.props.match.path}/service-profiles`} component={ListServiceProfiles} />
          <Route exact path={`${this.props.match.path}/service-profiles/create`} component={CreateServiceProfile} />
          <Route exact path={`${this.props.match.path}/service-profiles/:serviceProfileID`} component={UpdateServiceProfile} />
          <Route exact path={`${this.props.match.path}/device-profiles`} component={ListDeviceProfiles} />
          <Route match path={`${this.props.match.path}/device-profiles/create`} component={CreateDeviceProfile} />
          <Route match path={`${this.props.match.path}/device-profiles/:deviceProfileID`} component={UpdateDeviceProfile} />
        </Switch>
      </div>
    );
  }
}

export default withRouter(OrganizationLayout);
