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
import OrganizationStore from "../../stores/OrganizationStore";
import UpdateOrganization from "./UpdateOrganization";
import OrganizationDashboard from "./OrganizationDashboard";
import SessionStore from "../../stores/SessionStore";

import theme from "../../theme";


const styles = {
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "48px",
    overflow: "visible",
  },
};



class OrganizationLayout extends Component {
  constructor() {
    super();
    this.state = {
      tab: 0,
      admin: false,
    };
  }

  componentDidMount() {
    this.loadData();
    this.locationToTab();
    SessionStore.on("change", this.setIsAdmin);
    this.setIsAdmin();
  }

  componentWillUnmount() {
    SessionStore.removeListener("change", this.setIsAdmin);
  }

  componentDidUpdate(prevProps) {
    if (prevProps === this.props) {
      return;
    }

    this.loadData();
    this.locationToTab();
  }

  setIsAdmin = () => {
    this.setState({
      admin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID),
    });
  }

  loadData = () => {
    OrganizationStore.get(this.props.match.params.organizationID, resp => {
      this.setState({
        organization: resp,
      });
    });
  }

  deleteOrganization = () => {
    if (window.confirm("Are you sure you want to delete this organization?")) {
      OrganizationStore.delete(this.props.match.params.organizationID, () => {
        this.props.history.push("/organizations");
      });
    }
  }

  locationToTab = () => {
    let tab = 0;

    if (window.location.href.endsWith("/edit")) {
      tab = 1;
    } 

    this.setState({
      tab: tab,
    });
  }

  render() {
    if (this.state.organization === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={[
            <TitleBarButton
              key={1}
              label="Delete"
              icon={<Delete />}
              color="secondary"
              onClick={this.deleteOrganization}
            />,
          ]}
        >
          <TitleBarTitle to="/organizations" title="Organizations" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.organization.organization.name} />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs
            value={this.state.tab}
            indicatorColor="primary"
            className={this.props.classes.tabs}
          >
            <Tab label="Dashboard" component={Link} to={`/organizations/${this.props.match.params.organizationID}`} />
            <Tab label="Configuration" component={Link} to={`/organizations/${this.props.match.params.organizationID}/edit`} />
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={this.props.match.path} render={props => <OrganizationDashboard organization={this.state.organization.organization} {...props} />} />
            <Route exact path={`${this.props.match.path}/edit`} render={props => <UpdateOrganization organization={this.state.organization.organization} admin={this.state.admin} {...props} />} />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}


export default withStyles(styles)(withRouter(OrganizationLayout));
