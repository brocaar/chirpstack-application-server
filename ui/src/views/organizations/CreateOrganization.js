import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';

import { CardContent } from "@material-ui/core";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import OrganizationForm from "./OrganizationForm";
import OrganizationStore from "../../stores/OrganizationStore";


class CreateOrganization extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(organization) {
    OrganizationStore.create(organization, resp => {
      this.props.history.push("/organizations");
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Organizations" to="/organizations" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Create" />
        </TitleBar>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <OrganizationForm
                submitLabel="Create organization"
                onSubmit={this.onSubmit}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(CreateOrganization);