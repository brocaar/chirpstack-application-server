import React, { Component } from "react";

import Grid from '@material-ui/core/Grid';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

import Plus from "mdi-material-ui/Plus";
import HelpCircleOutline from "mdi-material-ui/HelpCircleOutline";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TableCellLink from "../../components/TableCellLink";
import TitleBarButton from "../../components/TitleBarButton";
import DataTable from "../../components/DataTable";

import GatewayProfileStore from "../../stores/GatewayProfileStore";


class ListGatewayProfiles extends Component {
  constructor() {
    super();

    this.state = {
      dialogOpen: false,
    };
  }

  getPage(limit, offset, callbackFunc) {
    GatewayProfileStore.list(0, limit, offset, callbackFunc);
  }

  getRow(obj) {
    return(
      <TableRow
        id={obj.id}
        hover
      >
        <TableCellLink to={`/gateway-profiles/${obj.id}`}>{obj.name}</TableCellLink>
        <TableCellLink to={`/network-servers/${obj.networkServerID}`}>{obj.networkServerName}</TableCellLink>
      </TableRow>
    );
  }

  toggleHelpDialog = () => {
    this.setState({
      dialogOpen: !this.state.dialogOpen,
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Dialog
          open={this.state.dialogOpen}
          onClose={this.toggleHelpDialog}
          aria-labelledby="help-dialog-title"
          aria-describedby="help-dialog-description"
        >
          <DialogTitle id="help-dialog-title">Gateway Profile help</DialogTitle>
          <DialogContent>
            <DialogContentText id="help-dialog-description">
              The only purpose of a Gateway Profile is to (re)configure one or multiple gateways with the
              configuration properties that are set by the Gateway Profile.<br /><br />

              When the ChirpStack Network Server detects that the configuration of a gateway is out-of-sync
              with its Gateway Profile, it will push a configuration command to the gateway in order to
              update its configuration.<br /><br />

              Please note that this feature is optional and only works in combination with the
              ChirpStack Concentratord component.<br /><br />

              Also note that the Gateway Profile does not change the way how devices are behaving.
              To configure the channel-plan that must be used by devices, update the
              ChirpStack Network Server configuration.
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={this.toggleHelpDialog} color="primary">Close</Button>
          </DialogActions>
        </Dialog>

        <TitleBar
          buttons={[
            <TitleBarButton
              key={1}
              label="Create"
              icon={<Plus />}
              to={`/gateway-profiles/create`}
            />,
            <TitleBarButton
              key={2}
              label="Help"
              icon={<HelpCircleOutline />}
              onClick={this.toggleHelpDialog}
            />
          ]}
        >
          <TitleBarTitle title="Gateway-profiles" />
        </TitleBar>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Network-server</TableCell>
              </TableRow>
            }
            getPage={this.getPage}
            getRow={this.getRow}
          />
        </Grid>
      </Grid>
    );
  }
}

export default ListGatewayProfiles;
