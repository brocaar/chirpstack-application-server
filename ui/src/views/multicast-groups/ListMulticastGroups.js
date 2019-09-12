import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

import Plus from "mdi-material-ui/Plus";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TableCellLink from "../../components/TableCellLink";
import TitleBarButton from "../../components/TitleBarButton";
import DataTable from "../../components/DataTable";
import DeviceAdmin from "../../components/DeviceAdmin";
import MulticastGroupStore from "../../stores/MulticastGroupStore";
import theme from "../../theme";


const styles = {
  idColumn: {
    width: theme.spacing(45),
    whiteSpace: "nowrap",
  },
};


class ListMulticastGroups extends Component {
  constructor() {
    super();
    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    MulticastGroupStore.list("", this.props.match.params.organizationID, "", "", limit, offset, callbackFunc);
  }

  getRow(obj) {
    return(
      <TableRow key={obj.id}>
        <TableCell>{obj.id}</TableCell>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/multicast-groups/${obj.id}`}>{obj.name}</TableCellLink>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/service-profiles/${obj.serviceProfileID}`}>{obj.serviceProfileName}</TableCellLink>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={
            <DeviceAdmin organizationID={this.props.match.params.organizationID}>
              <TitleBarButton
                label="Create"
                icon={<Plus />}
                to={`/organizations/${this.props.match.params.organizationID}/multicast-groups/create`}
              />
            </DeviceAdmin>
          }
        >
          <TitleBarTitle title="Multicast-groups" />
        </TitleBar>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell className={this.props.classes.idColumn}>ID</TableCell>
                <TableCell>Name</TableCell>
                <TableCell>Service-profile</TableCell>
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

export default withStyles(styles)(ListMulticastGroups);
