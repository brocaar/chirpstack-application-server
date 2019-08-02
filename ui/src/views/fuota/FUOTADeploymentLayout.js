import React, { Component } from "react";
import { Route, Switch, Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";

import ApplicationStore from "../../stores/ApplicationStore";
import FUOTADeploymentStore from "../../stores/FUOTADeploymentStore";
import FUOTADeploymentDetails from "./FUOTADeploymentDetails";
import ListFUOTADeploymentDevices from "./ListFUOTADeploymentDevices";

import theme from "../../theme";


const styles = {
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "48px",
    overflow: "visible",
  },
};


class FUOTADeploymentLayout extends Component {
  constructor() {
    super();

    this.state = {
      tab: 0,
    };

    this.onChangeTab = this.onChangeTab.bind(this);
    this.locationToTab = this.locationToTab.bind(this);
  }

  componentDidMount() {
    ApplicationStore.get(this.props.match.params.applicationID, resp => {
      this.setState({
        application: resp,
      });
    });

    FUOTADeploymentStore.on("reload", this.getFuotaDeployment);


    this.getFuotaDeployment();
    this.locationToTab();
  }

  componentWillUnmount() {
    FUOTADeploymentStore.removeListener("reload", this.getFuotaDeployment);
  }

  getFuotaDeployment = () => {
    FUOTADeploymentStore.get(this.props.match.params.fuotaDeploymentID, resp => {
      this.setState({
        fuotaDeployment: resp,
      });
    });
  }

  onChangeTab(e, v) {
    this.setState({
      tab: v,
    });
  }

  locationToTab() {
    let tab = 0;

    if (window.location.href.endsWith("/devices")) {
      tab = 1;
    }

    this.setState({
      tab: tab,
    });
  }


  render() {
    if (this.state.application === undefined || this.state.fuotaDeployment === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications`} title="Applications" />
          <TitleBarTitle title="/" />
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} title={this.state.application.application.name} />
          <TitleBarTitle title="/" />
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/fuota-deployments`} title="Firmware update jobs" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.fuotaDeployment.fuotaDeployment.name} />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs
            indicatorColor="primary"
            className={this.props.classes.tabs}
            value={this.state.tab}
            onChange={this.onChangeTab}
          >
            <Tab label="Information" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/fuota-deployments/${this.props.match.params.fuotaDeploymentID}`} />
            <Tab label="Devices" component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/fuota-deployments/${this.props.match.params.fuotaDeploymentID}/devices`} />
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={`${this.props.match.path}`} render={props => <FUOTADeploymentDetails fuotaDeployment={this.state.fuotaDeployment} {...props} />} />
            <Route exact path={`${this.props.match.path}/devices`} render={props => <ListFUOTADeploymentDevices fuotaDeployment={this.state.fuotaDeployment} {...props} />} />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(FUOTADeploymentLayout);

