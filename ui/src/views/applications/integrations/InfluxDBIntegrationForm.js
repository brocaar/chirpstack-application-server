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

  getVersionOptions(search, callbackFunc) {
    const versionOptions = [
      {value: "INFLUXDB_1", label: "InfuxDB 1.x"},
      {value: "INFLUXDB_2", label: "InfuxDB 2.x"},
    ];

    callbackFunc(versionOptions);
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <Form submitLabel={this.props.submitLabel} onSubmit={this.onSubmit}>
        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>InfluxDB version</FormLabel>
          <AutocompleteSelect
            id="version"
            label="Select InfluxDB version"
            value={this.state.object.version || ""}
            onChange={this.onChange}
            getOptions={this.getVersionOptions}
          />
        </FormControl>
        <TextField
          id="endpoint"
          label="API endpoint (write)"
          placeholder="http://localhost:8086/api/v2/write"
          value={this.state.object.endpoint || ""}
          onChange={this.onChange}
          margin="normal"
          required
          fullWidth
        />
        {this.state.object.version === "INFLUXDB_1" && <TextField
          id="username"
          label="Username"
          value={this.state.object.username || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
        />}
        {this.state.object.version === "INFLUXDB_1" && <TextField
          id="password"
          label="Password"
          value={this.state.object.password || ""}
          type="password"
          onChange={this.onChange}
          margin="normal"
          fullWidth
        />}
        {this.state.object.version === "INFLUXDB_1" && <TextField
          id="db"
          label="Database name"
          value={this.state.object.db || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />}
        {this.state.object.version === "INFLUXDB_1" && <TextField
          id="retentionPolicyName"
          label="Retention policy name"
          helperText="Sets the target retention policy for the write. InfluxDB writes to the DEFAULT retention policy if you do not specify a retention policy."
          value={this.state.object.retentionPolicyName || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
        />}
        {this.state.object.version === "INFLUXDB_1" && <FormControl fullWidth margin="normal">
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
        </FormControl>}
        {this.state.object.version === "INFLUXDB_2" && <TextField
          id="organization"
          label="Organization"
          value={this.state.object.organization || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />}
        {this.state.object.version === "INFLUXDB_2" && <TextField
          id="bucket"
          label="Bucket"
          value={this.state.object.bucket || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />}
        {this.state.object.version === "INFLUXDB_2" && <TextField
          id="token"
          label="Token"
          value={this.state.object.token || ""}
          type="password"
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />}
      </Form>
    );
  }
}

export default withStyles(styles)(InfluxDBIntegrationForm);
