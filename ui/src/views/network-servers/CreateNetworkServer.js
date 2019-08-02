import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';

import { CardContent } from "@material-ui/core";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import NetworkServerForm from "./NetworkServerForm";
import NetworkServerStore from "../../stores/NetworkServerStore";


class CreateNetworkServer extends Component {
  constructor() {
    super();
    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(networkServer) {
    NetworkServerStore.create(networkServer, resp => {
      this.props.history.push("/network-servers");
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Network-servers" to="/network-servers" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Add" />
        </TitleBar>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <NetworkServerForm
                submitLabel="Add network-server"
                onSubmit={this.onSubmit}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(CreateNetworkServer);
