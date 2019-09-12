import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Button from '@material-ui/core/Button';

import moment from "moment";
import Plus from "mdi-material-ui/Plus";
import PowerPlug from "mdi-material-ui/PowerPlug";

import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";
import DeviceAdmin from "../../components/DeviceAdmin";
import DeviceStore from "../../stores/DeviceStore";
import theme from "../../theme";


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


class ListDevices extends Component {
  constructor() {
    super();
    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    DeviceStore.list({
      applicationID: this.props.match.params.applicationID,
      limit: limit,
      offset: offset,
    }, callbackFunc);
  }

  getRow(obj) {
    let lastseen = "n/a";
    let margin = "n/a";
    let battery = "n/a";

    if (obj.lastSeenAt !== undefined && obj.lastSeenAt !== null) {
      lastseen = moment(obj.lastSeenAt).fromNow();
    }

    if (!obj.deviceStatusExternalPowerSource && !obj.deviceStatusBatteryLevelUnavailable) {
      battery = `${obj.deviceStatusBatteryLevel}%`
    }

    if (obj.deviceStatusExternalPowerSource) {
      battery = <PowerPlug />;
    }

    if (obj.deviceStatusMargin !== undefined && obj.deviceStatusMargin !== 256) {
      margin = `${obj.deviceStatusMargin} dB`;
    }

    return(
      <TableRow key={obj.devEUI}>
        <TableCell>{lastseen}</TableCell>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${obj.devEUI}`}>{obj.name}</TableCellLink>
        <TableCell>{obj.devEUI}</TableCell>
        <TableCell>{margin}</TableCell>
        <TableCell>{battery}</TableCell>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={4}>
        <DeviceAdmin organizationID={this.props.match.params.organizationID}>
          <Grid item xs={12} className={this.props.classes.buttons}>
            <Button variant="outlined" className={this.props.classes.button} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/create`}>
              <Plus className={this.props.classes.icon} />
              Create
            </Button>
          </Grid>
        </DeviceAdmin>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Last seen</TableCell>
                <TableCell>Device name</TableCell>
                <TableCell>Device EUI</TableCell>
                <TableCell>Link margin</TableCell>
                <TableCell>Battery</TableCell>
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

export default withStyles(styles)(ListDevices);
