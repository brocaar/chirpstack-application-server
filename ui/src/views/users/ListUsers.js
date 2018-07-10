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

import UserStore from "../../stores/UserStore";


class ListUsers extends Component {
  getPage(limit, offset, callbackFunc) {
    UserStore.list("", limit, offset, callbackFunc);
  }

  getRow(obj) {
    let active = null;
    let admin = null;

    if (obj.isAdmin) {
      admin = <Check />;
    } else {
      admin = <Close />;
    }

    if (obj.isActive) {
      active = <Check />;
    } else {
      active = <Close />;
    }

    return(
      <TableRow key={obj.id}>
        <TableCellLink to={`/users/${obj.id}`}>{obj.username}</TableCellLink>
        <TableCell>{active}</TableCell>
        <TableCell>{admin}</TableCell>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={24}>
        <TitleBar
          title="Users"
          buttons={[
            <TitleBarButton
              key={1}
              label="Create"
              icon={<Plus />}
              to={`/users/create`}
            />,
          ]}
        >
          <TitleBarTitle title="Users" />
        </TitleBar>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Username</TableCell>
                <TableCell>Active</TableCell>
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

export default ListUsers;
