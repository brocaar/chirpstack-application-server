import React, { Component } from "react";

import Grid from "@material-ui/core/Grid";
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import Table from "@material-ui/core/Table";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";

import moment from "moment";

import TableCellLink from "../../components/TableCellLink";


class DeviceDetails extends Component {
  render() {
    let lastSeenAt = "never";

    if (this.props.device.lastSeenAt !== null) {
      lastSeenAt = moment(this.props.device.lastSeenAt).format("lll");
    }


    return(
      <Grid container spacing={24}>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Details" />
            <CardContent>
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>{this.props.device.device.name}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Description</TableCell>
                    <TableCell>{this.props.device.device.description}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Device-profile</TableCell>
                    <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/device-profiles/${this.props.deviceProfile.deviceProfile.id}`}>{this.props.deviceProfile.deviceProfile.name}</TableCellLink>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Status" />
            <CardContent>
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell>Last seen at</TableCell>
                    <TableCell>{lastSeenAt}</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default DeviceDetails;
