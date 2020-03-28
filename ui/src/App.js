import React, { Component } from "react";
import {Router} from "react-router-dom";
import { Route, Switch } from 'react-router-dom';
import classNames from "classnames";

import CssBaseline from "@material-ui/core/CssBaseline";
import { MuiThemeProvider, withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';

import history from "./history";
import theme from "./theme";

import TopNav from "./components/TopNav";
import SideNav from "./components/SideNav";
import Footer from "./components/Footer";
import Notifications from "./components/Notifications";
import SessionStore from "./stores/SessionStore";

// network-server
import ListNetworkServers from "./views/network-servers/ListNetworkServers";
import CreateNetworkServer from "./views/network-servers/CreateNetworkServer";
import NetworkServerLayout from "./views/network-servers/NetworkServerLayout";

// gateway profiles
import ListGatewayProfiles from "./views/gateway-profiles/ListGatewayProfiles";
import CreateGatewayProfile from "./views/gateway-profiles/CreateGatewayProfile";
import GatewayProfileLayout from "./views/gateway-profiles/GatewayProfileLayout";

// organization
import ListOrganizations from "./views/organizations/ListOrganizations";
import CreateOrganization from "./views/organizations/CreateOrganization";
import OrganizationLayout from "./views/organizations/OrganizationLayout";
import ListOrganizationUsers from "./views/organizations/ListOrganizationUsers";
import OrganizationUserLayout from "./views/organizations/OrganizationUserLayout";
import CreateOrganizationUser from "./views/organizations/CreateOrganizationUser";
import OrganizationRedirect from "./views/organizations/OrganizationRedirect";

// user
import Login from "./views/users/Login";
import ListUsers from "./views/users/ListUsers";
import CreateUser from "./views/users/CreateUser";
import UserLayout from "./views/users/UserLayout";
import ChangeUserPassword from "./views/users/ChangeUserPassword";

// service-profile
import ListServiceProfiles from "./views/service-profiles/ListServiceProfiles";
import CreateServiceProfile from "./views/service-profiles/CreateServiceProfile";
import ServiceProfileLayout from "./views/service-profiles/ServiceProfileLayout";

// device-profile
import ListDeviceProfiles from "./views/device-profiles/ListDeviceProfiles";
import CreateDeviceProfile from "./views/device-profiles/CreateDeviceProfile";
import DeviceProfileLayout from "./views/device-profiles/DeviceProfileLayout";

// gateways
import ListGateways from "./views/gateways/ListGateways";
import GatewayLayout from "./views/gateways/GatewayLayout";
import CreateGateway from "./views/gateways/CreateGateway";

// applications
import ListApplications from "./views/applications/ListApplications";
import CreateApplication from "./views/applications/CreateApplication";
import ApplicationLayout from "./views/applications/ApplicationLayout";
import CreateIntegration from "./views/applications/CreateIntegration";
import UpdateIntegration from "./views/applications/UpdateIntegration";

// multicast-groups
import ListMulticastGroups from "./views/multicast-groups/ListMulticastGroups";
import CreateMulticastGroup from "./views/multicast-groups/CreateMulticastGroup";
import MulticastGroupLayout from "./views/multicast-groups/MulticastGroupLayout";
import AddDeviceToMulticastGroup from "./views/multicast-groups/AddDeviceToMulticastGroup";

// device
import CreateDevice from "./views/devices/CreateDevice";
import DeviceLayout from "./views/devices/DeviceLayout";

// search
import Search from "./views/search/Search";

// FUOTA
import CreateFUOTADeploymentForDevice from "./views/fuota/CreateFUOTADeploymentForDevice";
import FUOTADeploymentLayout from "./views/fuota/FUOTADeploymentLayout";


const drawerWidth = 270;

const styles = {
  root: {
    flexGrow: 1,
    display: "flex",
    minHeight: "100vh",
    flexDirection: "column",
  },
  paper: {
    padding: theme.spacing(2),
    textAlign: 'center',
    color: theme.palette.text.secondary,
  },
  main: {
    width: "100%",
    padding: 2 * 24,
    paddingTop: 115,
    flex: 1,
  },

  mainDrawerOpen: {
    paddingLeft: drawerWidth + (2 * 24),
  },
  footerDrawerOpen: {
    paddingLeft: drawerWidth,
  },
};


class App extends Component {
  constructor() {
    super();

    this.state = {
      user: null,
      drawerOpen: false,
    };

    this.setDrawerOpen = this.setDrawerOpen.bind(this);
  }

  componentDidMount() {
    SessionStore.on("change", () => {
      this.setState({
        user: SessionStore.getUser(),
        drawerOpen: SessionStore.getUser() != null,
      });
    });

    this.setState({
      user: SessionStore.getUser(),
      drawerOpen: SessionStore.getUser() != null,
    });
  }

  setDrawerOpen(state) {
    this.setState({
      drawerOpen: state,
    });
  }

  render() {
    let topNav = null;
    let sideNav = null;

    if (this.state.user !== null) {
      topNav = <TopNav setDrawerOpen={this.setDrawerOpen} drawerOpen={this.state.drawerOpen} user={this.state.user} />;
      sideNav = <SideNav open={this.state.drawerOpen} user={this.state.user} />
    }

    return (
      <Router history={history}>
        <React.Fragment>
          <CssBaseline />
          <MuiThemeProvider theme={theme}>
            <div className={this.props.classes.root}>
              {topNav}
              {sideNav}
              <div className={classNames(this.props.classes.main, this.state.drawerOpen && this.props.classes.mainDrawerOpen)}>
                <Grid container spacing={4}>
                  <Switch>
                    <Route exact path="/" component={OrganizationRedirect} />
                    <Route exact path="/login" component={Login} />
                    <Route exact path="/users" component={ListUsers} />
                    <Route exact path="/users/create" component={CreateUser} />
                    <Route exact path="/users/:userID(\d+)" component={UserLayout} />
                    <Route exact path="/users/:userID(\d+)/password" component={ChangeUserPassword} />

                    <Route exact path="/network-servers" component={ListNetworkServers} />
                    <Route exact path="/network-servers/create" component={CreateNetworkServer} />
                    <Route path="/network-servers/:networkServerID" component={NetworkServerLayout} />

                    <Route exact path="/gateway-profiles" component={ListGatewayProfiles} />
                    <Route exact path="/gateway-profiles/create" component={CreateGatewayProfile} />
                    <Route path="/gateway-profiles/:gatewayProfileID([\w-]{36})" component={GatewayProfileLayout} />

                    <Route exact path="/organizations/:organizationID(\d+)/service-profiles" component={ListServiceProfiles} />
                    <Route exact path="/organizations/:organizationID(\d+)/service-profiles/create" component={CreateServiceProfile} />
                    <Route path="/organizations/:organizationID(\d+)/service-profiles/:serviceProfileID([\w-]{36})" component={ServiceProfileLayout} />

                    <Route exact path="/organizations/:organizationID(\d+)/device-profiles" component={ListDeviceProfiles} />
                    <Route exact path="/organizations/:organizationID(\d+)/device-profiles/create" component={CreateDeviceProfile} />
                    <Route path="/organizations/:organizationID(\d+)/device-profiles/:deviceProfileID([\w-]{36})" component={DeviceProfileLayout} />

                    <Route exact path="/organizations/:organizationID(\d+)/gateways/create" component={CreateGateway} />
                    <Route path="/organizations/:organizationID(\d+)/gateways/:gatewayID([\w]{16})" component={GatewayLayout} />
                    <Route path="/organizations/:organizationID(\d+)/gateways" component={ListGateways} />

                    <Route exact path="/organizations/:organizationID(\d+)/applications" component={ListApplications} />
                    <Route exact path="/organizations/:organizationID(\d+)/applications/create" component={CreateApplication} />
                    <Route exact path="/organizations/:organizationID(\d+)/applications/:applicationID(\d+)/integrations/create" component={CreateIntegration} />
                    <Route exact path="/organizations/:organizationID(\d+)/applications/:applicationID(\d+)/integrations/:kind" component={UpdateIntegration} />
                    <Route exact path="/organizations/:organizationID(\d+)/applications/:applicationID(\d+)/devices/create" component={CreateDevice} />
                    <Route exact path="/organizations/:organizationID(\d+)/applications/:applicationID(\d+)/devices/:devEUI([\w]{16})/fuota-deployments/create" component={CreateFUOTADeploymentForDevice} />
                    <Route path="/organizations/:organizationID(\d+)/applications/:applicationID(\d+)/fuota-deployments/:fuotaDeploymentID([\w-]{36})" component={FUOTADeploymentLayout} />
                    <Route path="/organizations/:organizationID(\d+)/applications/:applicationID(\d+)/devices/:devEUI([\w]{16})" component={DeviceLayout} />
                    <Route path="/organizations/:organizationID(\d+)/applications/:applicationID(\d+)" component={ApplicationLayout} />

                    <Route exact path="/organizations/:organizationID(\d+)/multicast-groups" component={ListMulticastGroups} />
                    <Route exact path="/organizations/:organizationID(\d+)/multicast-groups/create" component={CreateMulticastGroup} />
                    <Route exact path="/organizations/:organizationID(\d+)/multicast-groups/:multicastGroupID/devices/create" component={AddDeviceToMulticastGroup} />
                    <Route path="/organizations/:organizationID(\d+)/multicast-groups/:multicastGroupID([\w-]{36})" component={MulticastGroupLayout} />

                    <Route exact path="/organizations" component={ListOrganizations} />
                    <Route exact path="/organizations/create" component={CreateOrganization} />
                    <Route exact path="/organizations/:organizationID(\d+)/users" component={ListOrganizationUsers} />
                    <Route exact path="/organizations/:organizationID(\d+)/users/create" component={CreateOrganizationUser} />
                    <Route exact path="/organizations/:organizationID(\d+)/users/:userID(\d+)" component={OrganizationUserLayout} />
                    <Route path="/organizations/:organizationID(\d+)" component={OrganizationLayout} />

                    <Route exact path="/search" component={Search} />
                  </Switch>
                </Grid>
              </div>
              <div className={this.state.drawerOpen ? this.props.classes.footerDrawerOpen : ""}>
                <Footer />
              </div>
            </div>
            <Notifications />
          </MuiThemeProvider>
        </React.Fragment>
      </Router>
    );
  }
}

export default withStyles(styles)(App);
