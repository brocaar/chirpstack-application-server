import React, { Component } from "react";

import Grid from "@material-ui/core/Grid";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';

import moment from "moment";

import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";

import FUOTADeploymentStore from "../../stores/FUOTADeploymentStore";


class FUOTADeploymentDevices extends Component {
  constructor() {
    super();

    this.state = {
      detailDialog: false,
    };

    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
    this.onCloseDialog = this.onCloseDialog.bind(this);
    this.showDetails = this.showDetails.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    FUOTADeploymentStore.listDeploymentDevices({
      fuota_deployment_id: this.props.match.params.fuotaDeploymentID,
      limit: limit,
      offset: offset,
    }, callbackFunc);
  }

  getRow(obj) {
    const createdAt = moment(obj.createdAt).format('lll');
    const updatedAt = moment(obj.updatedAt).format('lll');

    return(
      <TableRow key={obj.devEUI}>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${obj.devEUI}`}>{obj.deviceName}</TableCellLink>
        <TableCell>{obj.devEUI}</TableCell>
        <TableCell>{createdAt}</TableCell>
        <TableCell>{updatedAt}</TableCell>
        <TableCell><Button size="small" onClick={() => this.showDetails(obj.devEUI)}>{obj.state}</Button></TableCell>
      </TableRow>
    ); 
  }

  showDetails(devEUI) {
    FUOTADeploymentStore.getDeploymentDevice(this.props.match.params.fuotaDeploymentID, devEUI, resp => {
      this.setState({
        deploymentDevice: resp.deploymentDevice,
        detailDialog: true,
      });
    });
  }

  onCloseDialog() {
    this.setState({
      detailDialog: false,
    });
  }

  render() {
    let fddUpdatedAt = "";
    if (this.state.deploymentDevice !== undefined) {
      fddUpdatedAt = moment(this.state.deploymentDevice.updatedAt).format('lll');
    }

    return(
      <Grid container spacing={4}>
        {this.state.deploymentDevice && <Dialog
          open={this.state.detailDialog}
          onClose={this.onCloseDialog}
        >
          <DialogTitle>Job status for device</DialogTitle>
          <DialogContent>
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell>Last updated</TableCell>
                  <TableCell>{fddUpdatedAt}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>Device state</TableCell>
                  <TableCell>{this.state.deploymentDevice.state}</TableCell>
                </TableRow>
                {this.state.deploymentDevice.state === "ERROR" && <TableRow>
                  <TableCell>Error message</TableCell>
                  <TableCell>{this.state.deploymentDevice.errorMessage}</TableCell>
                </TableRow>}
              </TableBody>
            </Table>
          </DialogContent>
          <DialogActions>
            <Button color="primary" onClick={this.onCloseDialog}>Dismiss</Button>
          </DialogActions>
        </Dialog>}


        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Device name</TableCell>
                <TableCell>Device EUI</TableCell>
                <TableCell>Created at</TableCell>
                <TableCell>Updated at</TableCell>
                <TableCell>State</TableCell>
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

export default FUOTADeploymentDevices;
