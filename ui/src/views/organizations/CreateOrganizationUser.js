import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';
import TextField from "@material-ui/core/TextField";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import FormHelperText from "@material-ui/core/FormHelperText";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import AutocompleteSelect from "../../components/AutocompleteSelect";
import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import UserStore from "../../stores/UserStore";
import OrganizationStore from "../../stores/OrganizationStore";
import SessionStore from "../../stores/SessionStore";
import theme from "../../theme";


const styles = {
  card: {
    overflow: "visible",
  },
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "48px",
    overflow: "visible",
  },
  formLabel: {
    fontSize: 12,
  },
};



class AssignUserForm extends FormComponent {
  constructor() {
    super();

    this.getUserOptions = this.getUserOptions.bind(this);
  }

  getUserOptions(search, callbackFunc) {
    UserStore.list(search, 10, 0, resp => {
      const options = resp.result.map((u, i) => {return {label: u.username, value: u.id}});
      callbackFunc(options);
    });
  }

  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    return(
      <Form
        submitLabel="Add user"
        onSubmit={this.onSubmit}
      >
        <FormControl margin="normal" fullWidth>
          <FormLabel className={this.props.classes.formLabel} required>Username</FormLabel>
          <AutocompleteSelect
            id="userID"
            label="Select username"
            value={this.state.object.userID || null}
            onChange={this.onChange}
            getOptions={this.getUserOptions}
          />
        </FormControl>
        <Typography variant="body1">
          An user without additional permissions will be able to see all
          resources under this organization and will be able to send and
          receive device payloads.
        </Typography>
        <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="User is organization admin"
            control={
              <Checkbox
                id="isAdmin"
                checked={!!this.state.object.isAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
          <FormHelperText>An organization admin user is able to add and modify resources part of the organization.</FormHelperText>
        </FormControl>
        {!!!this.state.object.isAdmin && <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="User is device admin"
            control={
              <Checkbox
                id="isDeviceAdmin"
                checked={!!this.state.object.isDeviceAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
          <FormHelperText>A device admin user is able to add and modify resources part of the organization that are related to devices.</FormHelperText>
        </FormControl>}
        {!!!this.state.object.isAdmin && <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="User is gateway admin"
            control={
              <Checkbox
                id="isGatewayAdmin"
                checked={!!this.state.object.isGatewayAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
          <FormHelperText>A gateway admin user is able to add and modify gateways part of the organization.</FormHelperText>
        </FormControl>}
      </Form>
    );
  };
}

AssignUserForm = withStyles(styles)(AssignUserForm);


class CreateUserForm extends FormComponent {
  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    return(
      <Form
        submitLabel="Create user"
        onSubmit={this.onSubmit}
      >
        <TextField
          id="username"
          label="Username"
          margin="normal"
          value={this.state.object.username || ""}
          onChange={this.onChange}
          required
          fullWidth
        />
        <TextField
          id="email"
          label="E-mail address"
          margin="normal"
          value={this.state.object.email || ""}
          onChange={this.onChange}
          required
          fullWidth
        />
        <TextField
          id="note"
          label="Optional note"
          helperText="Optional note, e.g. a phone number, address, comment..."
          margin="normal"
          value={this.state.object.note || ""}
          onChange={this.onChange}
          rows={4}
          fullWidth
          multiline
        />
        <TextField
          id="password"
          label="Password"
          type="password"
          margin="normal"
          value={this.state.object.password || ""}
          onChange={this.onChange}
          required
          fullWidth
        />
        <Typography variant="body1">
          An user without additional permissions will be able to see all
          resources under this organization and will be able to send and
          receive device payloads.
        </Typography>
        <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="User is organization admin"
            control={
              <Checkbox
                id="isAdmin"
                checked={!!this.state.object.isAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
          <FormHelperText>An organization admin user is able to add and modify resources part of the organization.</FormHelperText>
        </FormControl>
        {!!!this.state.object.isAdmin && <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="User is device admin"
            control={
              <Checkbox
                id="isDeviceAdmin"
                checked={!!this.state.object.isDeviceAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
          <FormHelperText>A device admin user is able to add and modify resources part of the organization that are related to devices.</FormHelperText>
        </FormControl>}
        {!!!this.state.object.isAdmin && <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="User is gateway admin"
            control={
              <Checkbox
                id="isGatewayAdmin"
                checked={!!this.state.object.isGatewayAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
          <FormHelperText>A gateway admin user is able to add and modify gateways part of the organization.</FormHelperText>
        </FormControl>}
      </Form>
    );
  }
}


class CreateOrganizationUser extends Component {
  constructor() {
    super();

    this.state = {
      tab: 0,
      assignUser: false,
    };

    this.onChangeTab = this.onChangeTab.bind(this);
    this.onAssignUser = this.onAssignUser.bind(this);
    this.onCreateUser = this.onCreateUser.bind(this);
    this.setAssignUser = this.setAssignUser.bind(this);
  }

  componentDidMount() {
    this.setAssignUser();

    SessionStore.on("change", this.setAssignUser);
  }

  comomentWillUnmount() {
    SessionStore.removeListener("change", this.setAssignUser);
  }

  setAssignUser() {
    const settings = SessionStore.getSettings();
    this.setState({
      assignUser: !settings.disableAssignExistingUsers || SessionStore.isAdmin(),
    });
  }

  onChangeTab(e, v) {
    this.setState({
      tab: v,
    });
  }

  onAssignUser(user) {
    OrganizationStore.addUser(this.props.match.params.organizationID, user, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
    });
  };

  onCreateUser(user) {
    const orgs = [
      {isAdmin: user.isAdmin, isDeviceAdmin: user.isDeviceAdmin, isGatewayAdmin: user.isGatewayAdmin, organizationID: this.props.match.params.organizationID},
    ];

    let u = user;
    u.isActive = true;

    delete u.isAdmin;
    delete u.isDeviceAdmin;
    delete u.isGatewayAdmin;

    UserStore.create(u, user.password, orgs, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
    });
  };

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Organization users" to={`/organizations/${this.props.match.params.organizationID}/users`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Create" />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs value={this.state.tab} onChange={this.onChangeTab} indicatorColor="primary" className={this.props.classes.tabs}>
            {this.state.assignUser && <Tab label="Assign existing user" />}
            <Tab label="Create and assign user" />
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              {(this.state.tab === 0 && this.state.assignUser) && <AssignUserForm onSubmit={this.onAssignUser} />}
              {(this.state.tab === 1 || !this.state.assignUser) && <CreateUserForm onSubmit={this.onCreateUser} />}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(CreateOrganizationUser));
