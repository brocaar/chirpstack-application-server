import React, { Component } from "react";

import Paper from '@material-ui/core/Paper';
import { withStyles } from '@material-ui/core/styles';

import theme from "../theme";


const styles = {
  root: {
    padding: theme.spacing.unit * 2,
  },
};


class PaperComponent extends Component {
  render() {
    return(
      <Paper className={this.props.classes.root}>
        {this.props.children}
      </Paper>
    );
  }
}

export default withStyles(styles)(PaperComponent);