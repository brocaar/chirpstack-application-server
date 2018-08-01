import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Grid from '@material-ui/core/Grid';
import TextField from '@material-ui/core/TextField';
import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';
import CardContent from '@material-ui/core/CardContent';
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/core/styles";

import Form from "../../components/Form";
import FormComponent from "../../classes/FormComponent";
import SessionStore from "../../stores/SessionStore";
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
          id="username"
          label="Username"
          margin="normal"
          value={this.state.object.username || ""}
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


class Login extends Component {
  constructor() {
    super();

    this.state = {
      registration: null,
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    SessionStore.logout(() => {});

    SessionStore.getBranding(resp => {
      if (resp.registration !== "") {
        this.setState({
          registration: resp.registration,
        });
      }
    });
  }

  onSubmit(login) {
    SessionStore.login(login, () => {
      this.props.history.push("/");
    });
  }

  render() {
    return(
      <Grid container justify="center">
        <Grid item xs={6} lg={4}>
          <Card>
            <CardHeader
              title="Login"
            />
            <CardContent>
              <LoginForm
                submitLabel="Login"
                onSubmit={this.onSubmit}
              />
            </CardContent>
            {this.state.registration && <CardContent>
              <Typography className={this.props.classes.link} dangerouslySetInnerHTML={{__html: this.state.registration}}></Typography>
             </CardContent>}
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(Login));
