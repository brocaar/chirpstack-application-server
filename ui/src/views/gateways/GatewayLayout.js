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
import Admin from "../../components/Admin";
import GatewayStore from "../../stores/GatewayStore";
import SessionStore from "../../stores/SessionStore";
import GatewayDetails from "./GatewayDetails";
import UpdateGateway from "./UpdateGateway";
import GatewayDiscovery from "./GatewayDiscovery";
import GatewayFrames from "./GatewayFrames";

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


class GatewayLayout extends Component {
  constructor() {
    super();
    this.state = {
      tab: 0,
      wsStatus: null,
      admin: false,
    };
    this.deleteGateway = this.deleteGateway.bind(this);
    this.onChangeTab = this.onChangeTab.bind(this);
    this.locationToTab = this.locationToTab.bind(this);
    this.wsStatusUpdate = this.wsStatusUpdate.bind(this);
    this.setIsAdmin = this.setIsAdmin.bind(this);
  }

  componentDidMount() {
    GatewayStore.get(this.props.match.params.gatewayID, resp => {
      this.setState({
        gateway: resp,
      });
    });


    GatewayStore.on("ws.status.change", this.wsStatusUpdate);
    SessionStore.on("change", this.setIsAdmin);

    this.wsStatusUpdate();
    this.setIsAdmin();
    this.locationToTab();
  }

  componentDidUpdate(prevProps) {
    if (this.props === prevProps) {
      return;
    }

    this.locationToTab();
  }

  componentWillUnmount() {
    GatewayStore.removeListener("ws.status.change", this.wsStatusUpdate);
    SessionStore.removeListener("change", this.setIsAdmin);
  }

  setIsAdmin() {
    this.setState({
      admin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID),
    });
  }

  wsStatusUpdate() {
    this.setState({
      wsStatus: GatewayStore.getWSStatus(),
    });
  }

  deleteGateway() {
    if (window.confirm("Are you sure you want to delete this gateway?")) {
      GatewayStore.delete(this.props.match.params.gatewayID, () => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/gateways`);
      });
    }
  }

  locationToTab() {
    let tab = 0;

    if (window.location.href.endsWith("/edit")) {
      tab = 1;
    } else if (window.location.href.endsWith("/discovery")) {
      tab = 2;
    } else if (window.location.href.endsWith("/frames")) {
      tab = 3;
    }

    if (tab > 0 && !this.state.admin) {
      tab = tab - 1;
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
    if (this.state.gateway === undefined) {
      return(<div></div>);
    }

    let framesLabel = "Live LoRaWAN frames";

    if (this.state.wsStatus === "CONNECTED") {
      framesLabel = <Badge badgeContent="" className={this.props.classes.badgeGreen}>{framesLabel}</Badge>;
    } else if (this.state.tab === 3) {
      framesLabel = <Badge badgeContent="" color="error" className={this.props.classes.badge}>{framesLabel}</Badge>;
    }

    return(
      <Grid container spacing={24}>
        <TitleBar
          buttons={
            <Admin organizationID={this.props.match.params.organizationID}>
              <TitleBarButton
                key={1}
                label="Delete"
                icon={<Delete />}
                color="secondary"
                onClick={this.deleteGateway}
              />
            </Admin>
          }
        >
          <TitleBarTitle to={`/organizations/${this.props.match.params.organizationID}/gateways`} title="Gateways" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.gateway.gateway.name} />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs
            value={this.state.tab}
            onChange={this.onChangeTab}
            indicatorColor="primary"
            className={this.props.classes.tabs}
            fullWidth
          >
            <Tab label="Gateway details" component={Link} to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.gatewayID}`} />
            {this.state.admin && <Tab label="Gateway configuration" component={Link} to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.gatewayID}/edit`} />}
            <Tab label="Gateway discovery" disabled={!this.state.gateway.gateway.discoveryEnabled} component={Link} to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.gatewayID}/discovery`} />
            <Tab
              label={framesLabel}
              component={Link}
              to={`/organizations/${this.props.match.params.organizationID}/gateways/${this.props.match.params.gatewayID}/frames`}
            />
          </Tabs>
        </Grid>
        
        <Grid item xs={12}>
        <Switch>
          <Route exact path={`${this.props.match.path}`} render={props => <GatewayDetails gateway={this.state.gateway.gateway} lastSeenAt={this.state.gateway.lastSeenAt} {...props} />} />
          <Route exact path={`${this.props.match.path}/edit`} render={props => <UpdateGateway gateway={this.state.gateway.gateway} {...props} />} />
          <Route exact path={`${this.props.match.path}/discovery`} render={props => <GatewayDiscovery gateway={this.state.gateway.gateway} {...props} />} />
          <Route exact path={`${this.props.match.path}/frames`} render={props => <GatewayFrames gateway={this.state.gateway.gateway} {...props} />} />
        </Switch>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(GatewayLayout));
