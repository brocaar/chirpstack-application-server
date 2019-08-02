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
    let icon = null;

    if (obj.isAdmin) {
      icon = <Check />
    } else {
      icon = <Close />
    }

    return(
      <TableRow key={obj.userID}>
        <TableCell>{obj.userID}</TableCell>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/users/${obj.userID}`}>{obj.username}</TableCellLink>
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
                <TableCell>ID</TableCell>
                <TableCell>Username</TableCell>
                <TableCell>Admin</TableCell>
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
