import React, { Component } from "react";

import Table from '@material-ui/core/Table';
import TablePagination from '@material-ui/core/TablePagination';
import TableBody from '@material-ui/core/TableBody';
import TableHead from '@material-ui/core/TableHead';
import { withStyles } from '@material-ui/core/styles';

import Paper from "./Paper";


const styles = {

};


class DataTable extends Component {
  constructor() {
    super();

    this.state = {
      count: 0,
      rowsPerPage: 10,
      page: 0,
      loaded: {
        rows: false,
      },
    };

    this.onChangePage = this.onChangePage.bind(this);
    this.onChangeRowsPerPage = this.onChangeRowsPerPage.bind(this);
  }

  componentDidMount() {
    this.onChangePage(null, 0);
  }

  componentDidUpdate(prevProps) {
    if (this.props === prevProps) {
      return;
    }

    this.onChangePage(null, 0);
  }

  onChangePage(event, page) {
    this.props.getPage(this.state.rowsPerPage, (page) * this.state.rowsPerPage, (result) => {
      this.setState({
        page: page,
        count: parseInt(result.totalCount, 10),
        rows: result.result.map((row, i) => this.props.getRow(row)),
      });
    });
  }

  onChangeRowsPerPage(event) {
    this.setState({
      rowsPerPage: event.target.value,
    });

    this.props.getPage(event.target.value, 0, (result) => {
      this.setState({
        page: 0,
        count: result.totalCount,
        rows: result.result.map((row, i) => this.props.getRow(row)),
      });
    });
  }

  render() {
    if (this.state.rows === undefined) {
      return(<div></div>);
    }

    return(
      <Paper>
        <Table className={this.props.classes.table}>
          <TableHead>
            {this.props.header}
          </TableHead>
          <TableBody>
            {this.state.rows}
          </TableBody>
        </Table>
        <TablePagination
          component="div"
          count={this.state.count}
          rowsPerPage={this.state.rowsPerPage}
          page={this.state.page}
          onChangePage={this.onChangePage}
          onChangeRowsPerPage={this.onChangeRowsPerPage}
        />
      </Paper>
    );
  }
}

export default withStyles(styles)(DataTable);
