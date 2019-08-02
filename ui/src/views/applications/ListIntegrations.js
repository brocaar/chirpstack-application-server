import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Button from '@material-ui/core/Button';

import Plus from "mdi-material-ui/Plus";

import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";
import ApplicationStore from "../../stores/ApplicationStore";
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


class ListIntegrations extends Component {
  constructor() {
    super();
    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    ApplicationStore.listIntegrations(this.props.match.params.applicationID, callbackFunc);
  }

  getRow(obj) {
    const kind = obj.kind.toLowerCase();

    return(
      <TableRow key={obj.kind}>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations/${kind}`}>{obj.kind}</TableCellLink>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12} className={this.props.classes.buttons}>
          <Button variant="outlined" className={this.props.classes.button} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations/create`}>
            <Plus className={this.props.classes.icon} />
            Create
          </Button>
        </Grid>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Kind</TableCell>
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

export default withStyles(styles)(ListIntegrations);
