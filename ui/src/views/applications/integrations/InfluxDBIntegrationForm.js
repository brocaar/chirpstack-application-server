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


class InfluxDBIntegrationForm extends FormComponent {
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

    return(
      <Form submitLabel={this.props.submitLabel} onSubmit={this.onSubmit}>
        <TextField
          id="endpoint"
          label="API endpoint (write)"
          placeholder="http://localhost:8086/write"
          value={this.state.object.endpoint || ""}
          onChange={this.onChange}
          margin="normal"
          required
          fullWidth
        />
        <TextField
          id="username"
          label="Username"
          value={this.state.object.username || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
        />
        <TextField
          id="password"
          label="Password"
          value={this.state.object.password || ""}
          type="password"
          onChange={this.onChange}
          margin="normal"
          fullWidth
        />
        <TextField
          id="db"
          label="Database name"
          value={this.state.object.db || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />
        <TextField
          id="retentionPolicyName"
          label="Retention policy name"
          helperText="Sets the target retention policy for the write. InfluxDB writes to the DEFAULT retention policy if you do not specify a retention policy."
          value={this.state.object.retentionPolicyName || ""}
          onChange={this.onChange}
          margin="normal"
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

export default withStyles(styles)(InfluxDBIntegrationForm);
