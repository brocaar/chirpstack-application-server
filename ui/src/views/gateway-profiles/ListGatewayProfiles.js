import React, { Component } from "react";

import Grid from '@material-ui/core/Grid';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import Typography from "@material-ui/core/Typography";

import Plus from "mdi-material-ui/Plus";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TableCellLink from "../../components/TableCellLink";
import TitleBarButton from "../../components/TitleBarButton";
import DataTable from "../../components/DataTable";

import GatewayProfileStore from "../../stores/GatewayProfileStore";


class ListGatewayProfiles extends Component {
  getPage(limit, offset, callbackFunc) {
    GatewayProfileStore.list(0, limit, offset, callbackFunc);
  }

  getRow(obj) {
    return(
      <TableRow id={obj.id}>
        <TableCellLink to={`/gateway-profiles/${obj.id}`}>{obj.name}</TableCellLink>
        <TableCellLink to={`/network-servers/${obj.networkServerID}`}>{obj.networkServerName}</TableCellLink>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={[
            <TitleBarButton
              key={1}
              label="Create"
              icon={<Plus />}
              to={`/gateway-profiles/create`}
            />,
          ]}
        >
          <TitleBarTitle title="Gateway-profiles" />
        </TitleBar>
        <Grid item xs={12}>
              <Typography variant="subtitle2">
                Important: Gateway profiles are deprecated and will be removed in the next major release.
              </Typography>
        </Grid>
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
