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

import ApplicationStore from "../../stores/ApplicationStore";
import DeviceProfileStore from "../../stores/DeviceProfileStore";
import SessionStore from "../../stores/SessionStore";
import DeviceAdmin from "../../components/DeviceAdmin";
import DeviceStore from "../../stores/DeviceStore";
import UpdateDevice from "./UpdateDevice";
import DeviceKeys from "./DeviceKeys";
import DeviceActivation from "./DeviceActivation"
import DeviceData from "./DeviceData";
import DeviceFrames from "./DeviceFrames";
import DeviceDetails from "./DeviceDetails";

import theme from "../../theme";


const styles = {
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "49px",
  },
};


class DeviceLayout extends Component {
  constructor() {
    super();
    this.state = {
      tab: 0,
      admin: false,
    };

    this.onChangeTab = this.onChangeTab.bind(this);
    this.deleteDevice = this.deleteDevice.bind(this);
    this.locationToTab = this.locationToTab.bind(this);
    this.setIsAdmin = this.setIsAdmin.bind(this);
    this.getDevice = this.getDevice.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });


    DeviceStore.on("update", this.getDevice);
    SessionStore.on("change", this.setIsAdmin);

    this.locationToTab();
    this.setIsAdmin();
    this.getDevice();
  }

  componentWillUnmount() {
    SessionStore.removeListener("change", this.setIsAdmin);
    DeviceStore.removeListener("update", this.getDevice);
  }

  componentDidUpdate(oldProps) {
    if (this.props === oldProps) {
      return;
    }

    this.locationToTab();
  }

  getDevice() {
    DeviceStore.get(this.props.match.params.devEUI, resp => {
      this.setState({
        device: resp,
      });

      DeviceProfileStore.get(resp.device.deviceProfileID, resp => {
        this.setState({
          deviceProfile: resp,
        });
      });
    });
  }

  setIsAdmin() {
    this.setState({
      admin: SessionStore.isAdmin() || SessionStore.isOrganizationDeviceAdmin(this.props.match.params.organizationID),
    }, () => {
      // we need to update the tab index, as for non-admins, some tabs are hidden
      this.locationToTab();
    });
  }

  onChangeTab(e, v) {
    this.setState({
      tab: v,
    });
  }

  locationToTab() {
    let tab = 0;

    if (window.location.href.endsWith("/edit")) {
      tab = 1;
    } else if (window.location.href.endsWith("/keys")) {
      tab = 2;
    } else if (window.location.href.endsWith("/activation")) {
      tab = 3;
    } else if (window.location.href.endsWith("/data")) {
      tab = 4;
    } else if (window.location.href.endsWith("/frames")) {
      tab = 5;
    }

    if (tab > 1 && !this.state.admin) {
      tab = tab - 1;
    }

    this.setState({
      tab: tab,
    });
  }

  deleteDevice() {
    if (window.confirm("Are you sure you want to delete this device?")) {
      DeviceStore.delete(this.props.match.params.devEUI, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
      });
    }
  }

  render() {
    if (this.state.application === undefined || this.state.device === undefined|| this.state.deviceProfile === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={
            <DeviceAdmin organizationID={this.props.match.params.organizationID}>
              <TitleBarButton
                label="Delete"
                icon={<Delete />}
                color="secondary"
                onClick={this.deleteDevice}
              />
            </DeviceAdmin>
          }
        >
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications`} title="Applications" />
          <TitleBarTitle title="/" />
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} title={this.state.application.application.name} />
          <TitleBarTitle title="/" />
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} title="Devices" />
          <TitleBarTitle title="/" />
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}`} title={this.state.device.device.name} />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs
            value={this.state.tab}
            onChange={this.onChangeTab}
            indicatorColor="primary"
            className={this.props.classes.tabs}
                        variant="scrollable"
            scrollButtons="auto"
          >
            <Tab label="Details" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}`} />
            <Tab label="Configuration" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/edit`} />
            {this.state.admin && <Tab label="Keys (OTAA)" disabled={!this.state.deviceProfile.deviceProfile.supportsJoin} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/keys`} />}
            <Tab label="Activation" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/activation`} />
            <Tab label="Device data" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/data`} />
            <Tab label="LoRaWAN Frames" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/frames`} />
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={`${this.props.match.path}`} render={props => <DeviceDetails device={this.state.device} deviceProfile={this.state.deviceProfile} admin={this.state.admin} {...props} />} />
            <Route exact path={`${this.props.match.path}/edit`} render={props => <UpdateDevice device={this.state.device.device} admin={this.state.admin} {...props} />} />
            <Route exact path={`${this.props.match.path}/keys`} render={props => <DeviceKeys device={this.state.device.device} admin={this.state.admin} deviceProfile={this.state.deviceProfile.deviceProfile} {...props} />} />
            <Route exact path={`${this.props.match.path}/activation`} render={props => <DeviceActivation device={this.state.device.device} admin={this.state.admin} deviceProfile={this.state.deviceProfile.deviceProfile} {...props} />} />
            <Route exact path={`${this.props.match.path}/data`} render={props => <DeviceData device={this.state.device.device} admin={this.state.admin} {...props} />} />
            <Route exact path={`${this.props.match.path}/frames`} render={props => <DeviceFrames device={this.state.device.device} admin={this.state.admin} {...props} />} />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(DeviceLayout));
