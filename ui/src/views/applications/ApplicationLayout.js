import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';

import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";
import Admin from "../../components/Admin";

import ApplicationStore from "../../stores/ApplicationStore";
import SessionStore from "../../stores/SessionStore";
import ListDevices from "../devices/ListDevices";
import UpdateApplication from "./UpdateApplication";
import ListIntegrations from "./ListIntegrations";

import ListMulticastGroups from "../multicast-groups/ListMulticastGroups";

import CreateAWSSNSIntegration from "./integrations/CreateAWSSNSIntegration";
import CreateGCPPubSubIntegration from "./integrations/CreateGCPPubSubIntegration";
import CreateHTTPIntegration from "./integrations/CreateHTTPIntegration";
import CreateAzureServiceBusIntegration from "./integrations/CreateAzureServiceBusIntegration";
import CreateInfluxDBIntegration from "./integrations/CreateInfluxDBIntegration";
import CreateThingsBoardIntegration from "./integrations/CreateThingsBoardIntegration";
import CreateLoRaCloudIntegration from "./integrations/CreateLoRaCloudIntegration";
import CreateMyDevicesIntegration from "./integrations/CreateMyDevicesIntegration";
import CreatePilotThingsIntegration from "./integrations/CreatePilotThingsIntegration";
import UpdateAWSSNSIntegration from "./integrations/UpdateAWSSNSIntegration";
import UpdateGCPPubSubIntegration from "./integrations/UpdateGCPPubSubIntegration";
import UpdateHTTPIntegration from "./integrations/UpdateHTTPIntegration";
import UpdateAzureServiceBusIntegration from "./integrations/UpdateAzureServiceBusIntegration";
import UpdateInfluxDBIntegration from "./integrations/UpdateInfluxDBIntegration";
import UpdateThingsBoardIntegration from "./integrations/UpdateThingsBoardIntegration";
import UpdateLoRaCloudIntegration from "./integrations/UpdateLoRaCloudIntegration";
import UpdateMyDevicesIntegration from "./integrations/UpdateMyDevicesIntegration";
import UpdatePilotThingsIntegration from "./integrations/UpdatePilotThingsIntegration";
import MQTTCertificate from "./integrations/MQTTCertificate";

import theme from "../../theme";


const styles = {
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "48px",
    overflow: "visible",
  },
};


class ApplicationLayout extends Component {
  constructor() {
    super();
    this.state = {
      tab: 0,
      admin: false,
    };

    this.deleteApplication = this.deleteApplication.bind(this);
    this.locationToTab = this.locationToTab.bind(this);
    this.onChangeTab = this.onChangeTab.bind(this);
    this.setIsAdmin = this.setIsAdmin.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });

    SessionStore.on("change", this.setIsAdmin);

    this.setIsAdmin();
    this.locationToTab();
  }

  componentWillUnmount() {
    SessionStore.removeListener("change", this.setIsAdmin);
  }

  componentDidUpdate(oldProps) {
    if (this.props === oldProps) {
      return;
    }

    this.locationToTab();
  }

  setIsAdmin() {
    this.setState({
      admin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID),
    });
  }

  deleteApplication() {
    if (window.confirm("Are you sure you want to delete this application? This will also delete all devices part of this application.")) {
      ApplicationStore.delete(this.props.match.params.applicationID, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications`);
      });
    }
  }

  locationToTab() {
    let tab = 0;

    if (window.location.href.match(/.*\/integrations.*/g)) {
      tab = 3;
    } else if (window.location.href.endsWith("/edit")) {
      tab = 2;
    } else if (window.location.href.match(/.*\/multicast-groups.*/g)) {
      tab = 1;
    }

    this.setState({
      tab: tab,
    });
  }

  onChangeTab(e, v) {
    this.setState({
      tab: v,
    });
  }

  render() {
    if (this.state.application === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={
            <Admin organizationID={this.props.match.params.organizationID}>
              <TitleBarButton
                label="Delete"
                icon={<Delete />}
                color="secondary"
                onClick={this.deleteApplication}
              />
            </Admin>
          }
        >
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications`} title="Applications" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.application.application.name} />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs
            value={this.state.tab}
            onChange={this.onChangeTab}
            indicatorColor="primary"
            className={this.props.classes.tabs}
          >
            <Tab label="Devices" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
            <Tab label="Multicast groups" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups`} />
            {this.state.admin && <Tab label="Application configuration" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/edit`} />}
            {this.state.admin && <Tab label="Integrations" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`} />}
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={`${this.props.match.path}/multicast-groups`} render={props => <ListMulticastGroups application={this.state.application.application} {...props} />} />
            <Route exact path={`${this.props.match.path}/edit`} render={props => <UpdateApplication application={this.state.application.application} {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations`} render={props => <ListIntegrations application={this.state.application.application} {...props} />} />
            <Route exact path={`${this.props.match.path}`} render={props => <ListDevices application={this.state.application.application} {...props} />} />

            <Route exact path={`${this.props.match.path}/integrations/aws-sns/create`} render={props => <CreateAWSSNSIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/aws-sns/edit`} render={props => <UpdateAWSSNSIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/gcp-pubsub/create`} render={props => <CreateGCPPubSubIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/gcp-pubsub/edit`} render={props => <UpdateGCPPubSubIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/http/create`} render={props => <CreateHTTPIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/http/edit`} render={props => <UpdateHTTPIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/azure-service-bus/create`} render={props => <CreateAzureServiceBusIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/azure-service-bus/edit`} render={props => <UpdateAzureServiceBusIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/influxdb/create`} render={props => <CreateInfluxDBIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/influxdb/edit`} render={props => <UpdateInfluxDBIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/thingsboard/create`} render={props => <CreateThingsBoardIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/thingsboard/edit`} render={props => <UpdateThingsBoardIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/loracloud/create`} render={props => <CreateLoRaCloudIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/loracloud/edit`} render={props => <UpdateLoRaCloudIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/mydevices/create`} render={props => <CreateMyDevicesIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/mydevices/edit`} render={props => <UpdateMyDevicesIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/pilot-things/create`} render={props => <CreatePilotThingsIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/pilot-things/edit`} render={props => <UpdatePilotThingsIntegration {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations/mqtt/certificate`} render={props => <MQTTCertificate {...props} />} />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(ApplicationLayout));
