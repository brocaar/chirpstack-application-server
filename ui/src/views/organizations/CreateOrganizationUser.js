import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import OrganizationStore from "../../stores/OrganizationStore";
import OrganizationUserForm from "./OrganizationUserForm";


class CreateOrganizationUser extends Component {
  constructor() {
    super();

    this.onAssignUser = this.onAssignUser.bind(this);
  }

  onAssignUser(user) {
    OrganizationStore.addUser(this.props.match.params.organizationID, user, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/users`);
    });
  };

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Organization users" to={`/organizations/${this.props.match.params.organizationID}/users`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Add" />
        </TitleBar>

        <Grid item xs={12}>
          <Card>
            <CardContent>
              <OrganizationUserForm
                submitLabel="Add user"
                onSubmit={this.onAssignUser}
                update={false}
              />
            </CardContent>
          </Card>

        </Grid>
      </Grid>
    );
  }
}

export default withRouter(CreateOrganizationUser);
