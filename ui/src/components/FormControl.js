import React, { Component } from "react";

import FormControl from '@material-ui/core/FormControl';
import FormLabel from '@material-ui/core/FormLabel';
import { withStyles } from "@material-ui/core/styles";

import theme from "../theme";


const styles = {
  formControl: {
    marginTop: theme.spacing.unit * 4,
  },
  formLabel: {
    color: theme.palette.primary.main,
  },
};


class FormControlComponent extends Component {
  render() {
    return(
      <FormControl margin="normal" className={this.props.classes.formControl} fullWidth={true}>
        <FormLabel className={this.props.classes.formLabel}>
          {this.props.label}
        </FormLabel>
        {this.props.children}
      </FormControl>
    );
  }
}

export default withStyles(styles)(FormControlComponent);
