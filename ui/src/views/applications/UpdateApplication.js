import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import ApplicationStore from "../../stores/ApplicationStore";
import ApplicationForm from "./ApplicationForm";


const styles = {
  card: {
    overflow: "visible",
  },
};


class UpdateApplication extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(application) {
    ApplicationStore.update(application, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${application.id}`);
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <ApplicationForm
                submitLabel="Update application"
                object={this.props.application}
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

export default withStyles(styles)(withRouter(UpdateApplication));
