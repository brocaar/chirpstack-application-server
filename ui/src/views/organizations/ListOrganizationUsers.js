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


class ListOrganizationUsers extends Component {
  constructor() {
    super();
    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    OrganizationStore.listUsers(this.props.match.params.organizationID, limit, offset, callbackFunc);
  }

  getRow(obj) {
    let orgAdmin = null;
    let gwAdmin = null;
    let devAdmin = null;

    if (obj.isAdmin) {
      orgAdmin = <Check />
    } else {
      orgAdmin = <Close />
    }

    if (obj.isAdmin || obj.isGatewayAdmin) {
      gwAdmin = <Check />
    } else {
      gwAdmin = <Close />
    }

    if (obj.isAdmin || obj.isDeviceAdmin) {
      devAdmin = <Check />
    } else {
      devAdmin = <Close />
    }

    return(
      <TableRow
        key={obj.userID}
        hover
      >
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/users/${obj.userID}`}>{obj.email}</TableCellLink>
        <TableCell>{orgAdmin}</TableCell>
        <TableCell>{gwAdmin}</TableCell>
        <TableCell>{devAdmin}</TableCell>
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
              label="Add"
              icon={<Plus />}
              to={`/organizations/${this.props.match.params.organizationID}/users/create`}
            />,
          ]}
        >
          <TitleBarTitle title="Organization users" />
        </TitleBar>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Email</TableCell>
                <TableCell>Organization admin</TableCell>
                <TableCell>Gateway admin</TableCell>
                <TableCell>Device admin</TableCell>
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

export default ListOrganizationUsers;
