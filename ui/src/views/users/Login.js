import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Grid from '@material-ui/core/Grid';
import TextField from '@material-ui/core/TextField';
import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';
import CardContent from '@material-ui/core/CardContent';
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import { withStyles } from "@material-ui/core/styles";

import queryString from "query-string";

import Form from "../../components/Form";
import FormComponent from "../../classes/FormComponent";
import SessionStore from "../../stores/SessionStore";
import InternalStore from "../../stores/InternalStore";
import theme from "../../theme";


const styles = {
  textField: {
    width: "100%",
  },
  link: {
    "& a": {
      color: theme.palette.primary.main,
      textDecoration: "none",
    },
  },
};


class LoginForm extends FormComponent {
  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <TextField
          id="email"
          label="Username / email"
          margin="normal"
          value={this.state.object.email || ""}
          onChange={this.onChange}
          fullWidth
          required
        />
        <TextField
          id="password"
          label="Password"
          type="password"
          margin="normal"
          value={this.state.object.password || ""}
          onChange={this.onChange}
          fullWidth
          required
        />
      </Form>
    );
  }
}

class OpenIDConnectLogin extends Component {
  render() {
    return(
      <div>
        <a href={this.props.loginUrl}><Button variant="outlined">{this.props.loginLabel}</Button></a>
      </div>
    );
  }
}


class Login extends Component {
  constructor() {
    super();

    this.state = {
      loaded: false,
      registration: "",
      oidcEnabled: false,
      oidcLoginlabel: "",
      oidcLoginUrl: "",
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    SessionStore.logout(true, () => {});

    InternalStore.settings(resp => {
      this.setState({
        loaded: true,
        registration: resp.branding.registration,
        oidcEnabled: resp.openidConnect.enabled,
        oidcLoginUrl: resp.openidConnect.loginURL,
        oidcLoginLabel: resp.openidConnect.loginLabel,
      });
    });

    // callback from openid provider
    if (this.props.location.search !== "") {
      let query = queryString.parse(this.props.location.search);
  
      SessionStore.openidConnectLogin(query.code, query.state, () => {
        this.props.history.push("/");
      });
    }
  }

  onSubmit(login) {
    SessionStore.login(login, () => {
      this.props.history.push("/");
    });
  }

  render() {
    if (!this.state.loaded) {
      return null;
    }

    return(
      <Grid container justify="center">
        <Grid item xs={6} lg={4}>
          <Card>
            <CardHeader
              title="ChirpStack Login"
            />
            <CardContent>
              {!this.state.oidcEnabled && <LoginForm
                submitLabel="Login"
                onSubmit={this.onSubmit}
              />}

              {this.state.oidcEnabled && <OpenIDConnectLogin
                loginUrl={this.state.oidcLoginUrl}
                loginLabel={this.state.oidcLoginLabel}
              />}
            </CardContent>
            {this.state.registration !== "" && <CardContent>
              <Typography className={this.props.classes.link} dangerouslySetInnerHTML={{__html: this.state.registration}}></Typography>
             </CardContent>}
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(Login));
