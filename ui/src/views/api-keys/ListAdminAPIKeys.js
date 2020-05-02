import React, { Component } from "react";
import { withRouter } from "react-router-dom";

import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import IconButton from '@material-ui/core/IconButton';

import Plus from "mdi-material-ui/Plus";
import Delete from "mdi-material-ui/Delete";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TitleBarButton from "../../components/TitleBarButton";
import DataTable from "../../components/DataTable";
import InternalStore from "../../stores/InternalStore";


class APIKeyRow extends Component {
  onDelete = () => {
    if (window.confirm("Are you sure you want to delete this api key?")) {
      InternalStore.deleteAPIKey(this.props.obj.id, resp => {
        this.props.history.push(`/api-keys`);
      });
    }
  }

  render() {
    return(
      <TableRow
        key={this.props.obj.id}
        hover
      >
        <TableCell>{this.props.obj.id}</TableCell>
        <TableCell>{this.props.obj.name}</TableCell>
        <TableCell align="right">
          <IconButton aria-label="delete"><Delete onClick={this.onDelete} /></IconButton>
        </TableCell>
      </TableRow>
    );
  }
}

APIKeyRow = withRouter(APIKeyRow);


class ListAdminAPIKeys extends Component {
  getPage = (limit, offset, callbackFunc) => {
    InternalStore.listAPIKeys({
      isAdmin: true,
      limit: limit,
      offset: offset,
    }, callbackFunc);
  }

  getRow = (obj) => {
    return(<APIKeyRow obj={obj} />);
  }

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={
            <TitleBarButton
              label="Create"
              icon={<Plus />}
              to={`/api-keys/create`}
            />
          }
        >
          <TitleBarTitle title="Global API keys" />
        </TitleBar>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Name</TableCell>
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

export default ListAdminAPIKeys;
