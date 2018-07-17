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
import FormGroup from "@material-ui/core/FormGroup";
import Checkbox from '@material-ui/core/Checkbox';
import TextField from "@material-ui/core/TextField";
import CardContent from "@material-ui/core/CardContent";

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

    this.getUserOption = this.getUserOption.bind(this);
    this.getUserOptions = this.getUserOptions.bind(this);
  }

  getUserOption(id, callbackFunc) {
    UserStore.get(id, resp => {
      callbackFunc({label: resp.user.username, value: resp.user.id});
    });
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
            getOption={this.getUserOption}
            getOptions={this.getUserOptions}
          />
        </FormControl>
        <FormGroup>
          <FormControlLabel
            label="Is organization admin"
            control={
              <Checkbox
                id="isAdmin"
                checked={!!this.state.object.isAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
        </FormGroup>
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
        <FormGroup>
          <FormControlLabel
            label="Is organization admin"
            control={
              <Checkbox
                id="isAdmin"
                checked={!!this.state.object.isAdmin}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
        </FormGroup>
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
      {isAdmin: user.isAdmin, organizationID: this.props.match.params.organizationID},
    ];

    let u = user;
    u.isAdmin = false;
    u.isActive = true;

    UserStore.create(u, user.password, orgs, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
    });
  };

  render() {
    return(
      <Grid container spacing={24}>
        <TitleBar>
          <TitleBarTitle title="Organization users" to={`/organizations/${this.props.match.params.organizationID}/users`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Create" />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs value={this.state.tab} onChange={this.onChangeTab} indicatorColor="primary" fullWidth className={this.props.classes.tabs}>
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
