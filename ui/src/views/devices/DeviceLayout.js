import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import Badge from '@material-ui/core/Badge';

import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";

import ApplicationStore from "../../stores/ApplicationStore";
import DeviceProfileStore from "../../stores/DeviceProfileStore";
import SessionStore from "../../stores/SessionStore";
import Admin from "../../components/Admin";
import DeviceStore from "../../stores/DeviceStore";
import UpdateDevice from "./UpdateDevice";
import DeviceKeys from "./DeviceKeys";
import DeviceActivation from "./DeviceActivation"
import DeviceData from "./DeviceData";
import DeviceFrames from "./DeviceFrames";

import theme from "../../theme";


const styles = {
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "48px",
    overflow: "visible",
  },
  badge: {
    padding: `0 ${theme.spacing.unit * 2}px`,
  },
  badgeGreen: {
    padding: `0 ${theme.spacing.unit * 2}px`,
    "& span": {
      backgroundColor: "#4CAF50 !important",
    },
  },
};


class DeviceLayout extends Component {
  constructor() {
    super();
    this.state = {
      tab: 0,
      admin: false,
      wsDataStatus: null,
      wsFramesStatus: null,
    };

    this.onChangeTab = this.onChangeTab.bind(this);
    this.deleteDevice = this.deleteDevice.bind(this);
    this.locationToTab = this.locationToTab.bind(this);
    this.wsStatusChange = this.wsStatusChange.bind(this);
    this.setIsAdmin = this.setIsAdmin.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });

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

    DeviceStore.on("ws.status.change", this.wsStatusChange);
    SessionStore.on("change", this.setIsAdmin);

    this.locationToTab();
    this.setIsAdmin();
  }

  componentWillUnmount() {
    SessionStore.removeListener("change", this.setIsAdmin);
    DeviceStore.on("ws.status.change", this.wsStatusChange);
  }

  setIsAdmin() {
    this.setState({
      admin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID),
    }, () => {
      // we need to update the tab index, as for non-admins, some tabs are hidden
      this.locationToTab();
    });
  }

  wsStatusChange() {
    this.setState({
      wsDataStatus: DeviceStore.getWSDataStatus(),
      wsFramesStatus: DeviceStore.getWSFramesStatus(),
    });
  }

  onChangeTab(e, v) {
    this.setState({
      tab: v,
    });
  }

  locationToTab() {
    let tab = 0;

    if (window.location.href.endsWith("/keys")) {
      tab = 1;
    } else if (window.location.href.endsWith("/activation")) {
      tab = 2;
    } else if (window.location.href.endsWith("/data")) {
      tab = 3;
    } else if (window.location.href.endsWith("/frames")) {
      tab = 4;
    }

    if (tab > 0 && !this.state.admin) {
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

    let dataLabel = "Live device data";
    let framesLabel = "Live LoRaWAN frames";

    if (this.state.wsDataStatus === "CONNECTED") {
      dataLabel = <Badge badgeContent="" className={this.props.classes.badgeGreen}>{dataLabel}</Badge>;
    } else if ((this.state.admin && this.state.tab === 3) || (!this.state.admin && this.state.tab === 2)) {
      dataLabel = <Badge badgeContent="" color="error" className={this.props.classes.badge}>{dataLabel}</Badge>;
    }

    if (this.state.wsFramesStatus === "CONNECTED") {
      framesLabel = <Badge badgeContent="" className={this.props.classes.badgeGreen}>{framesLabel}</Badge>;
    } else if ((this.state.admin && this.state.tab === 4) || (!this.state.admin && this.state.tab === 3)) {
      framesLabel = <Badge badgeContent="" color="error" className={this.props.classes.badge}>{framesLabel}</Badge>;
    }

    return(
      <Grid container spacing={24}>
        <TitleBar
          buttons={
            <Admin organizationID={this.props.match.params.organizationID}>
              <TitleBarButton
                label="Delete"
                icon={<Delete />}
                color="secondary"
                onClick={this.deleteDevice}
              />
            </Admin>
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
            fullWidth
          >
            <Tab label="Configuration" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}`} />
            {this.state.admin && <Tab label="Keys (OTAA)" disabled={!this.state.deviceProfile.deviceProfile.supportsJoin} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/keys`} />}
            <Tab label="Activation" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/activation`} />
            <Tab label={dataLabel} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/data`} />
            <Tab label={framesLabel} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}/frames`} />
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={`${this.props.match.path}`} render={props => <UpdateDevice device={this.state.device.device} admin={this.state.admin} {...props} />} />
            <Route exact path={`${this.props.match.path}/keys`} render={props => <DeviceKeys device={this.state.device.device} admin={this.state.admin} {...props} />} />
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
