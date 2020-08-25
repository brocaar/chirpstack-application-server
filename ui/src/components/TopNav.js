import React, { Component } from "react";
import { withRouter, Link } from 'react-router-dom';

import AppBar from "@material-ui/core/AppBar";
import Toolbar from "@material-ui/core/Toolbar";
import { withStyles } from "@material-ui/core/styles";
import { IconButton } from "@material-ui/core";
import MenuItem from '@material-ui/core/MenuItem';
import Menu from '@material-ui/core/Menu';
import Input from "@material-ui/core/Input";
import InputAdornment from "@material-ui/core/InputAdornment";
import blue from "@material-ui/core/colors/blue";
import Avatar from '@material-ui/core/Avatar';
import Chip from '@material-ui/core/Chip';

import MenuIcon from "mdi-material-ui/Menu";
import Backburger from "mdi-material-ui/Backburger";
import AccountCircle from "mdi-material-ui/AccountCircle";
import Magnify from "mdi-material-ui/Magnify";
import HelpCicle from "mdi-material-ui/HelpCircle";

import InternalStore from "../stores/InternalStore";
import SessionStore from "../stores/SessionStore";
import theme from "../theme";


const styles = {
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
  },
  menuButton: {
    marginLeft: -12,
    marginRight: 10,
  },
  hidden: {
    display: "none",
  },
  flex: {
    flex: 1,
  },
  logo: {
    height: 32,
  },
  search: {
    marginRight: 3 * theme.spacing(1),
    color: theme.palette.common.white,
    background: blue[400],
    width: 450,
    padding: 5,
    borderRadius: 3,
  },
  chip: {
    background: blue[600],
    color: theme.palette.common.white,
    marginRight: theme.spacing(1),
    "&:hover": {
      background: blue[400],
    },
    "&:active": {
      background: blue[400],
    },
  },
  iconButton: {
    color: theme.palette.common.white,
    marginRight: theme.spacing(1),
  },
};


class TopNav extends Component {
  constructor() {
    super();

    this.state = {
      menuAnchor: null,
      search: "",
      oidcEnabled: false,
    };
  }

  onMenuOpen = (e) => {
    this.setState({
      menuAnchor: e.currentTarget,
    });
  }

  onMenuClose = () => {
    this.setState({
      menuAnchor: null,
    });
  }

  onLogout = () => {
    if (this.state.oidcEnabled === true) {
      if (this.state.logoutURL !== "") {
        SessionStore.logout(false, () => {
          window.location.assign(this.state.logoutURL);
        });
      } else {
        SessionStore.logout(true, () => {
            this.props.history.push("/login");
        });
      }
    } else {
      SessionStore.logout(true, () => {
        this.props.history.push("/login");
      });
    }
  }

  handleDrawerToggle = () => {
    this.props.setDrawerOpen(!this.props.drawerOpen);
  }

  onSearchChange = (e) => {
    this.setState({
      search: e.target.value,
    });
  }

  onSearchSubmit = (e) => {
    e.preventDefault();
    this.props.history.push(`/search?search=${encodeURIComponent(this.state.search)}`);
  }

  componentDidMount() {
    InternalStore.settings(resp => {
      this.setState({
        oidcEnabled: resp.openidConnect.enabled,
        logoutURL: resp.openidConnect.logoutURL,
      });
    })
  }

  render() {
    let drawerIcon;
    if (!this.props.drawerOpen) {
      drawerIcon = <MenuIcon />;
    } else {
      drawerIcon = <Backburger />;
    }

    const open = Boolean(this.state.menuAnchor);

    return(
      <AppBar className={this.props.classes.appBar}>
        <Toolbar>
          <IconButton
            color="inherit"
            aria-label="toggle drawer"
            onClick={this.handleDrawerToggle}
            className={this.props.classes.menuButton}
          >
            {drawerIcon}
          </IconButton>

          <div className={this.props.classes.flex}>
            <img src="/logo/logo.png" className={this.props.classes.logo} alt="ChirpStack.io" />
          </div>

          <form onSubmit={this.onSearchSubmit}>
            <Input
              placeholder="Search organization, application, gateway or device"
              className={this.props.classes.search}
              disableUnderline={true}
              value={this.state.search || ""}
              onChange={this.onSearchChange}
              startAdornment={
                <InputAdornment position="start">
                  <Magnify />
                </InputAdornment>
              }
            />
          </form>

          <a href="https://www.chirpstack.io/application-server/" target="chirpstack-doc">
            <IconButton className={this.props.classes.iconButton}>
              <HelpCicle />
            </IconButton>
          </a>

          <Chip
            avatar={
              <Avatar>
                <AccountCircle />
              </Avatar>
            }
            label={this.props.user.email}
            onClick={this.onMenuOpen}
            color="primary"
            classes={{
              root: this.props.classes.chip,
            }}
          />
          <Menu
            id="menu-appbar"
            anchorEl={this.state.menuAnchor}
            anchorOrigin={{
              vertical: "top",
              horizontal: "right",
            }}
            transformOrigin={{
              vertical: "top",
              horizontal: "right",
            }}
            open={open}
            onClose={this.onMenuClose}
          >
            {!this.state.oidcEnabled && <MenuItem component={Link} to={`/users/${this.props.user.id}/password`}>Change password</MenuItem>}
            <MenuItem onClick={this.onLogout}>Logout</MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>
    );
  }
}

export default withStyles(styles)(withRouter(TopNav));
