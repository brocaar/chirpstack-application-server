import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

import moment from "moment";

import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";
import FUOTADeploymentStore from "../../stores/FUOTADeploymentStore";
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


class ListFUOTADeploymentsForApplication extends Component {
  constructor() {
    super();

    this.state = {};

    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    FUOTADeploymentStore.list({
      applicationID: this.props.match.params.applicationID,
      limit: limit,
      offset: offset,
    }, callbackFunc);
  }

  getRow(obj) {
    const createdAt = moment(obj.createdAt).format('lll');
    const updatedAt = moment(obj.updatedAt).format('lll');

    return(
      <TableRow key={obj.id}>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/fuota-deployments/${obj.id}`}>{obj.name}</TableCellLink>
        <TableCell>{createdAt}</TableCell>
        <TableCell>{updatedAt}</TableCell>
        <TableCell>{obj.state}</TableCell>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Created at</TableCell>
                <TableCell>Updated at</TableCell>
                <TableCell>State</TableCell>
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

export default withStyles(styles)(ListFUOTADeploymentsForApplication);

