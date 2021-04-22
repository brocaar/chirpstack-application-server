import React from "react";

import { withStyles } from "@material-ui/core/styles";
import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import TextField from '@material-ui/core/TextField';
import FormHelperText from "@material-ui/core/FormHelperText";

import FormComponent from "../../../classes/FormComponent";
import Form from "../../../components/Form";
import AutocompleteSelect from "../../../components/AutocompleteSelect";


const styles = {
  formLabel: {
    fontSize: 12,
  },
};


class InfluxDBV2IntegrationForm extends FormComponent {
  getPrecisionOptions(search, callbackFunc) {
    const precisionOptions = [
      {value: "NS", label: "Nanosecond"},
      {value: "U", label: "Microsecond"},
      {value: "MS", label: "Millisecond"},
      {value: "S", label: "Second"},
      {value: "M", label: "Minute"},
      {value: "H", label: "Hour"},
    ];

    callbackFunc(precisionOptions);
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    // onSubmit={this.onSubmit}>
    return(
      <Form submitLabel={this.props.submitLabel} onSubmit={this.onSubmit} >
        <TextField
          id="host"
          label="Host"
          placeholder="http://localhost:8086"
          value={this.state.object.host || ""}
          onChange={this.onChange}
          margin="normal"
          required
          fullWidth
        />
        <TextField
          id="org"
          label="Organisation"
          value={this.state.object.org || ""}
          onChange={this.onChange}
          margin="normal"
          required
          fullWidth
        />
        <TextField
          id="bucket"
          label="Bucket"
          value={this.state.object.bucket || ""}
          onChange={this.onChange}
          margin="normal"
          required
          fullWidth
        />
        <TextField
          id="token"
          label="Token"
          value={this.state.object.token || ""}
          onChange={this.onChange}
          margin="normal"
          type="password"
          required
          fullWidth
        />
        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Timestamp precision</FormLabel>
          <AutocompleteSelect
            id="precision"
            label="Select timestamp precision"
            value={this.state.object.precision || ""}
            onChange={this.onChange}
            getOptions={this.getPrecisionOptions}
          />
          <FormHelperText>
            It is recommented to use the least precise precision possible as this can result in significant improvements in compression.
          </FormHelperText>
        </FormControl>
      </Form>
    );
  }
}

export default withStyles(styles)(InfluxDBV2IntegrationForm);
