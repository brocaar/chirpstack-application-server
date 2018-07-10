import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import { CardContent } from "@material-ui/core";

import UserStore from "../../stores/UserStore";
import UserForm from "./UserForm";

class UpdateUser extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(user) {
    UserStore.update(user, resp => {
      this.props.history.push("/users");
    });
  }

  render() {
    return(
      <Grid container spacing={24}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <UserForm
                submitLabel="Update user"
                object={this.props.user}
                onSubmit={this.onSubmit}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(UpdateUser);
