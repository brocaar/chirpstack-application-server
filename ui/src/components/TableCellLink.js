import React, { Component } from "react";
import { Link } from 'react-router-dom';

import TableCell from '@material-ui/core/TableCell';
import { withStyles } from '@material-ui/core/styles';

import theme from "../theme";


const styles = {
  link: {
    textDecoration: "none",
    color: theme.palette.primary.main,
  },
};


class TableCellLink extends Component {
  render() {
    return(
      <TableCell>
        <Link className={this.props.classes.link} to={this.props.to}>{this.props.children}</Link>
      </TableCell>
    );
  }
}

export default withStyles(styles)(TableCellLink);
