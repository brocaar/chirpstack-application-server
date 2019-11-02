import React, { Component } from "react";
import { withRouter, Link } from 'react-router-dom';

import { withStyles } from '@material-ui/core/styles';
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import { CardContent } from "@material-ui/core";
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import Button from "@material-ui/core/Button";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import GatewayProfileForm from "./GatewayProfileForm";
import GatewayProfileStore from "../../stores/GatewayProfileStore";
import NetworkServerStore from "../../stores/NetworkServerStore";


const styles = {
  card: {
    overflow: "visible",
  },
};


class CreateGatewayProfile extends Component {
  constructor() {
    super();
    this.state = {
      nsDialog: false,
    };
    this.onSubmit = this.onSubmit.bind(this);
    this.closeDialog = this.closeDialog.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.list(0, 0, 0, resp => {
      if (resp.totalCount === "0") {
        this.setState({
          nsDialog: true,
        });
      }
    });
  }

  closeDialog() {
    this.setState({
      nsDialog: false,
    });
  }

  onSubmit(gatewayProfile) {
    GatewayProfileStore.create(gatewayProfile, resp => {
      this.props.history.push("/gateway-profiles");
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Dialog
          open={this.state.nsDialog}
          onClose={this.closeDialog}
        >
          <DialogTitle>Add a network-server?</DialogTitle>
          <DialogContent>
            <DialogContentText paragraph>
              ChirpStack Application Server isn't connected to a ChirpStack Network Server network-server.
              Did you know that ChirpStack Application Server can connect to multiple ChirpStack Network Server instances, e.g. to support multiple regions?
            </DialogContentText>
            <DialogContentText>
              Would you like to connect to a network-server now?
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button color="primary" component={Link} to="/network-servers/create" onClick={this.closeDialog}>Add network-server</Button>
            <Button color="primary" onClick={this.closeDialog}>Dismiss</Button>
          </DialogActions>
        </Dialog>

        <TitleBar>
          <TitleBarTitle title="Gateway-profiles" to="/gateway-profiles" />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Create" />
        </TitleBar>

        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <GatewayProfileForm
                submitLabel="Create gateway-profile"
                onSubmit={this.onSubmit}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(CreateGatewayProfile));
