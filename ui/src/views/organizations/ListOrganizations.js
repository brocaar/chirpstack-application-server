import React, { Component } from "react";

import Grid from '@material-ui/core/Grid';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';

import Check from "mdi-material-ui/Check";
import Close from "mdi-material-ui/Close";
import Plus from "mdi-material-ui/Plus";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TableCellLink from "../../components/TableCellLink";
import TitleBarButton from "../../components/TitleBarButton";
import DataTable from "../../components/DataTable";

import OrganizationStore from "../../stores/OrganizationStore";


class ListOrganizations extends Component {
  getPage(limit, offset, callbackFunc) {
    OrganizationStore.list("", limit, offset, callbackFunc);
  }

  getRow(obj) {
    let icon = null;

    if (obj.canHaveGateways) {
      icon = <Check />;
    } else {
      icon = <Close />;
    }

    return(
      <TableRow
        key={obj.id}
        hover
      >
        <TableCellLink to={`/organizations/${obj.id}`}>{obj.name}</TableCellLink>
        <TableCell>{obj.displayName}</TableCell>
        <TableCell>{icon}</TableCell>
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
              to={`/organizations/create`}
            />,
          ]}
        >
          <TitleBarTitle title="Organizations" />
        </TitleBar>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Display name</TableCell>
                <TableCell>Can have gateways</TableCell>
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

export default ListOrganizations;
