import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import MulticastGroupStore from "../../stores/MulticastGroupStore";
import MulticastGroupForm from "./MulticastGroupForm";


const styles = {
  card: {
    overflow: "visible",
  },
};


class UpdateMulticastGroup extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(multicastGroup) {
    MulticastGroupStore.update(multicastGroup, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups/${this.props.match.params.multicastGroupID}`);
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <MulticastGroupForm
                submitLabel="Update multicast-group"
                object={this.props.multicastGroup}
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

export default withStyles(styles)(withRouter(UpdateMulticastGroup));
