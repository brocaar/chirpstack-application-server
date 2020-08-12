import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import OrganzationStore from "../../stores/OrganizationStore";
import OrganizationForm from "./OrganizationForm";


class UpdateOrganization extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(organization) {
    OrganzationStore.update(organization, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}`);
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <OrganizationForm
                submitLabel="Update organization"
                object={this.props.organization}
                onSubmit={this.onSubmit}
                disabled={!this.props.admin}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(UpdateOrganization);
