import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import OrganizationStore from "../../stores/OrganizationStore";
import OrganizationUserForm from "./OrganizationUserForm";


class UpdateOrganizationUser extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(organizationUser) {
    OrganizationStore.updateUser(organizationUser, resp => {
      this.props.history.push(`/organizations/${organizationUser.organizationID}/users`);
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <OrganizationUserForm
                submitLabel="Update user"
                object={this.props.organizationUser}
                onSubmit={this.onSubmit}
                update={true}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(UpdateOrganizationUser);
