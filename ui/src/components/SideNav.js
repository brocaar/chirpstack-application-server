import React, { Component } from "react";
import { Link, withRouter } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';

import Divider from '@material-ui/core/Divider';
import Domain from "mdi-material-ui/Domain";
import Home from "mdi-material-ui/Home";
import Account from "mdi-material-ui/Account";
import Server from "mdi-material-ui/Server";
import Apps from "mdi-material-ui/Apps";
import RadioTower from "mdi-material-ui/RadioTower";
import Tune from "mdi-material-ui/Tune";
import AccountDetails from "mdi-material-ui/AccountDetails";
import KeyVariant from "mdi-material-ui/KeyVariant";

import AutocompleteSelect from "./AutocompleteSelect";
import SessionStore from "../stores/SessionStore";
import OrganizationStore from "../stores/OrganizationStore";
import Admin from "./Admin";

import theme from "../theme";


const styles = {
  drawerPaper: {
    position: "fixed",
    width: 270,
    paddingTop: theme.spacing(9),
  },
  select: {
    paddingTop: theme.spacing(1),
    paddingLeft: theme.spacing(3),
    paddingRight: theme.spacing(3),
    paddingBottom: theme.spacing(1),
  },
};

class SideNav extends Component {
  constructor() {
    super();

    this.state = {
      open: true,
      organization: null,
      cacheCounter: 0,
    };


    this.onChange = this.onChange.bind(this);
    this.getOrganizationOption = this.getOrganizationOption.bind(this);
    this.getOrganizationOptions = this.getOrganizationOptions.bind(this);
    this.getOrganizationFromLocation = this.getOrganizationFromLocation.bind(this);
  }

  componentDidMount() {
    SessionStore.on("organization.change", () => {
      OrganizationStore.get(SessionStore.getOrganizationID(), resp => {
        this.setState({
          organization: resp.organization,
        });
      });
    });

    OrganizationStore.on("create", () => {
      this.setState({
        cacheCounter: this.state.cacheCounter + 1,
      });
    });

    OrganizationStore.on("change", (org) => {
      if (this.state.organization !== null && this.state.organization.id === org.id) {
        this.setState({
          organization: org,
        });
      }

      this.setState({
        cacheCounter: this.state.cacheCounter + 1,
      });
    });

    OrganizationStore.on("delete", id => {
      if (this.state.organization !== null && this.state.organization.id === id) {
        this.setState({
          organization: null,
        });
      }

      this.setState({
        cacheCounter: this.state.cacheCounter + 1,
      });
    });

    if (SessionStore.getOrganizationID() !== null) {
      OrganizationStore.get(SessionStore.getOrganizationID(), resp => {
        this.setState({
          organization: resp.organization,
        });
      });
    }

    this.getOrganizationFromLocation();
  }

  componentDidUpdate(prevProps) {
    if (this.props === prevProps) {
      return;
    }

    this.getOrganizationFromLocation();
  }

  onChange(e) {
    this.props.history.push(`/organizations/${e.target.value}`);
  }

  getOrganizationFromLocation() {
    const organizationRe = /\/organizations\/(\d+)/g;
    const match = organizationRe.exec(this.props.history.location.pathname);

    if (match !== null && (this.state.organization === null || this.state.organization.id !== match[1])) {
      SessionStore.setOrganizationID(match[1]);
    }
  }

  getOrganizationOption(id, callbackFunc) {
    OrganizationStore.get(id, resp => {
      callbackFunc({label: resp.organization.name, value: resp.organization.id});
    });
  }

  getOrganizationOptions(search, callbackFunc) {
    OrganizationStore.list(search, 10, 0, resp => {
      const options = resp.result.map((o, i) => {return {label: o.name, value: o.id}});
      callbackFunc(options, resp.totalCount);
    });
  }

  render() {
    let organizationID = "";
    if (this.state.organization !== null) {
      organizationID = this.state.organization.id;
    }

    return(
      <Drawer
        variant="persistent"
        anchor="left"
        open={this.props.open}
        classes={{paper: this.props.classes.drawerPaper}}
      >
        <Admin>
          <List>
            <ListItem button component={Link} to="/dashboard">
              <ListItemIcon>
                <Home />
              </ListItemIcon>
              <ListItemText primary="Dashboard" />
            </ListItem>
            <ListItem button component={Link} to="/network-servers">
              <ListItemIcon>
                <Server />
              </ListItemIcon>
              <ListItemText primary="Network-servers" />
            </ListItem>
            <ListItem button component={Link} to="/gateway-profiles">
              <ListItemIcon>
                <RadioTower />
              </ListItemIcon>
              <ListItemText primary="Gateway-profiles" />
            </ListItem>
            <ListItem button component={Link} to="/organizations">
            <ListItemIcon>
                <Domain />
              </ListItemIcon>
              <ListItemText primary="Organizations" />
            </ListItem>
            <ListItem button component={Link} to="/users">
              <ListItemIcon>
                <Account />
              </ListItemIcon>
              <ListItemText primary="All users" />
            </ListItem>
            <ListItem button component={Link} to="/api-keys">
              <ListItemIcon>
                <KeyVariant />
              </ListItemIcon>
              <ListItemText primary="API keys" />
            </ListItem>
          </List>
          <Divider />
        </Admin>

        <div>
          <AutocompleteSelect
            id="organizationID"
            margin="none"
            value={organizationID}
            onChange={this.onChange}
            getOption={this.getOrganizationOption}
            getOptions={this.getOrganizationOptions}
            className={this.props.classes.select}
            triggerReload={this.state.cacheCounter}
          />
        </div>

        {this.state.organization && <List>
          <ListItem button component={Link} to={`/organizations/${this.state.organization.id}`}>
            <ListItemIcon>
              <Home />
            </ListItemIcon>
            <ListItemText primary="Org. dashboard" />
          </ListItem>
          <Admin organizationID={this.state.organization.id}>
            <ListItem button component={Link} to={`/organizations/${this.state.organization.id}/users`}>
              <ListItemIcon>
                <Account />
              </ListItemIcon>
              <ListItemText primary="Org. users" />
            </ListItem>
            <ListItem button component={Link} to={`/organizations/${this.state.organization.id}/api-keys`}>
              <ListItemIcon>
                <KeyVariant />
              </ListItemIcon>
              <ListItemText primary="Org. API keys" />
            </ListItem>
          </Admin>
          <ListItem button component={Link} to={`/organizations/${this.state.organization.id}/service-profiles`}>
            <ListItemIcon>
              <AccountDetails />
            </ListItemIcon>
            <ListItemText primary="Service-profiles" />
          </ListItem>
          <ListItem button component={Link} to={`/organizations/${this.state.organization.id}/device-profiles`}>
            <ListItemIcon>
              <Tune />
            </ListItemIcon>
            <ListItemText primary="Device-profiles" />
          </ListItem>
          {this.state.organization.canHaveGateways && <ListItem button component={Link} to={`/organizations/${this.state.organization.id}/gateways`}>
            <ListItemIcon>
              <RadioTower />
            </ListItemIcon>
            <ListItemText primary="Gateways" />
          </ListItem>}
          <ListItem button component={Link} to={`/organizations/${this.state.organization.id}/applications`}>
            <ListItemIcon>
              <Apps />
            </ListItemIcon>
            <ListItemText primary="Applications" />
          </ListItem>
        </List>}
      </Drawer>
    );
  }
}

export default withRouter(withStyles(styles)(SideNav));
