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
import ListFUOTADeploymentsForApplication from "../fuota/ListFUOTADeploymentsForApplication";


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

    if (window.location.href.endsWith("/edit")) {
      tab = 1;
    } else if (window.location.href.endsWith("/integrations")) {
      tab = 2;
    } else if (window.location.href.endsWith("/fuota-deployments")) {
      tab = 3;
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
            {this.state.admin && <Tab label="Application configuration" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/edit`} />}
            {this.state.admin && <Tab label="Integrations" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`} />}
            {this.state.admin && <Tab label="FUOTA" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/fuota-deployments`} />}
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={`${this.props.match.path}/edit`} render={props => <UpdateApplication application={this.state.application.application} {...props} />} />
            <Route exact path={`${this.props.match.path}/integrations`} render={props => <ListIntegrations application={this.state.application.application} {...props} />} />
            <Route exact path={`${this.props.match.path}`} render={props => <ListDevices application={this.state.application.application} {...props} />} />
            <Route exact path={`${this.props.match.path}/fuota-deployments`} render={props => <ListFUOTADeploymentsForApplication application={this.state.application.application} {...props} /> } />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(ApplicationLayout));
