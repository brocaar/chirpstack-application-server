import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Button from '@material-ui/core/Button';
import IconButton from "@material-ui/core/IconButton";

import Plus from "mdi-material-ui/Plus";
import Delete from "mdi-material-ui/Delete";

import DeviceAdmin from "../../components/DeviceAdmin";
import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";
import DeviceStore from "../../stores/DeviceStore";
import theme from "../../theme";
import multicastGroupStore from "../../stores/MulticastGroupStore";


const styles = {
  buttons: {
    textAlign: "right",
  },
  button: {
    marginLeft: 2 * theme.spacing(1),
  },
  icon: {
    marginRight: theme.spacing(1),
  },
};



class ListMulticastGroupDevices extends Component {
  constructor() {
    super();
    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    DeviceStore.list({
      multicastGroupID: this.props.match.params.multicastGroupID,
      limit: limit,
      offset: offset,
    }, callbackFunc);
  }

  onDelete(devEUI) {
    if(window.confirm("Are you sure you want to remove this device from the multicast-group? This does not remove the device itself.")) {
      multicastGroupStore.removeDevice(this.props.match.params.multicastGroupID, devEUI, resp => {
        this.forceUpdate();
      });
    }
  }

  getRow(obj) {
    return(
      <TableRow key={obj.devEUI}>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${obj.applicationID}/devices/${obj.devEUI}`}>{obj.name}</TableCellLink>
        <TableCell>{obj.devEUI}</TableCell>
        <TableCell className={this.props.classes.buttons}>
          <DeviceAdmin organizationID={this.props.match.params.organizationID}>
            <IconButton onClick={this.onDelete.bind(this, obj.devEUI)}><Delete /></IconButton>
          </DeviceAdmin>
        </TableCell>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={4}>
        <DeviceAdmin organizationID={this.props.match.params.organizationID}>
          <Grid item xs={12} className={this.props.classes.buttons}>
            <Button variant="outlined" className={this.props.classes.button} component={Link} to={`/organizations/${this.props.match.params.organizationID}/multicast-groups/${this.props.match.params.multicastGroupID}/devices/create`}>
              <Plus className={this.props.classes.icon} />
              Add
            </Button>
          </Grid>
        </DeviceAdmin>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Device name</TableCell>
                <TableCell>Device EUI</TableCell>
                <TableCell></TableCell>
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

export default withStyles(styles)(ListMulticastGroupDevices);
