import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import { CardContent } from "@material-ui/core";

import NetworkServerStore from "../../stores/NetworkServerStore";
import NetworkServerForm from "./NetworkServerForm";


class UpdateNetworkServer extends Component {
  constructor() {
    super();

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(networkServer) {
    NetworkServerStore.update(networkServer, resp => {
      this.props.history.push("/network-servers");
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <NetworkServerForm
                submitLabel="Update network-server"
                object={this.props.networkServer}
                onSubmit={this.onSubmit}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(UpdateNetworkServer);
