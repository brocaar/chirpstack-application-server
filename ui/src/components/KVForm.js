import React from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import IconButton from '@material-ui/core/IconButton';
import TextField from '@material-ui/core/TextField';

import Delete from "mdi-material-ui/Delete";

import FormComponent from "../classes/FormComponent";
import theme from "../theme";


const kvStyles = {
  formLabel: {
    fontSize: 12,
  },
  delete: {
    marginTop: 3 * theme.spacing(1),
  },
};


class KVForm extends FormComponent {
  onChange(e) {
    super.onChange(e);
    this.props.onChange(this.props.index, this.state.object);
  }

  onDelete = (e) => {
    e.preventDefault();
    this.props.onDelete(this.props.index);
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <Grid item xs={4}>
          <TextField
            id="key"
            label="Name"
            margin="normal"
            value={this.state.object.key || ""}
            onChange={this.onChange}
            disabled={!!this.props.disabled}
            fullWidth
          />
        </Grid>
        <Grid item xs={7}>
          <TextField
            id="value"
            label="Value"
            margin="normal"
            value={this.state.object.value || ""}
            onChange={this.onChange}
            disabled={!!this.props.disabled}
            fullWidth
          />
        </Grid>
        <Grid item xs={1} className={this.props.classes.delete}>
          {!!!this.props.disabled && <IconButton aria-label="delete" onClick={this.onDelete}>
            <Delete />
          </IconButton>}
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(kvStyles)(KVForm);
