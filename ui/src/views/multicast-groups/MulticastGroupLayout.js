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
import DeviceAdmin from "../../components/DeviceAdmin";

import ApplicationStore from "../../stores/ApplicationStore";
import MulticastGroupStore from "../../stores/MulticastGroupStore";
import SessionStore from "../../stores/SessionStore";
import UpdateMulticastGroup from "./UpdateMulticastGroup";
import theme from "../../theme";
import ListMulticastGroupDevices from "./ListMulticastGroupDevices";


const styles = {
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "48px",
    overflow: "visible",
  },
};


class MulticastGroupLayout extends Component {
  constructor() {
    super();
    this.state = {
      tab: 0,
      admin: false,
    };

    this.locationToTab = this.locationToTab.bind(this);
    this.onChangeTab = this.onChangeTab.bind(this);
    this.deleteMulticastGroup = this.deleteMulticastGroup.bind(this);
    this.setIsAdmin = this.setIsAdmin.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });

    MulticastGroupStore.get(this.props.match.params.multicastGroupID, resp => {
      this.setState({
        multicastGroup: resp,
      });
    });

    SessionStore.on("change", this.setIsAdmin);
    this.setIsAdmin();
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
      admin: SessionStore.isAdmin() || SessionStore.isOrganizationDeviceAdmin(this.props.match.params.organizationID),
    });
  }

  locationToTab() {
    let tab = 0;

    if (window.location.href.endsWith("/edit")) {
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

  deleteMulticastGroup() {
    if (window.confirm("Are you sure you want to delete this multicast-group?")) {
      MulticastGroupStore.delete(this.props.match.params.multicastGroupID, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups`);
      });
    }
  }

  render() {
    if (this.state.application === undefined || this.state.multicastGroup === undefined) {
      return null;
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
                onClick={this.deleteMulticastGroup}
              />
            </DeviceAdmin>
          }
        >
          <TitleBarTitle title="Applications" to={`/organizations/${this.props.match.params.organizationID}/applications`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.application.application.name} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Multicast groups" to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.multicastGroup.multicastGroup.name} />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs
            value={this.state.tab}
            onChange={this.onChangeTab}
            indicatorColor="primary"
            className={this.props.classes.tabs}
          >
            <Tab label="Devices" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups/${this.props.match.params.multicastGroupID}`} />
            {this.state.admin && <Tab label="Configuration" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups/${this.props.match.params.multicastGroupID}/edit`} />}
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={`${this.props.match.path}/edit`} render={props => <UpdateMulticastGroup multicastGroup={this.state.multicastGroup.multicastGroup} {...props} />} />
            <Route exact path={`${this.props.match.path}`} render={props => <ListMulticastGroupDevices multicastGroup={this.state.multicastGroup.multicastGroup} {...props} />} />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(MulticastGroupLayout));
