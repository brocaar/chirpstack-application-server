import React from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TextField from '@material-ui/core/TextField';
import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import IconButton from '@material-ui/core/IconButton';
import FormHelperText from "@material-ui/core/FormHelperText";
import Button from "@material-ui/core/Button";

import Delete from "mdi-material-ui/Delete";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import AutocompleteSelect from "../../components/AutocompleteSelect";
import theme from "../../theme";


const styles = {
  delete: {
    marginTop: 3 * theme.spacing(1),
  },
  formLabel: {
    fontSize: 12,
  },
};


class HTTPIntegrationHeaderForm extends FormComponent {
  constructor() {
    super();

    this.onDelete = this.onDelete.bind(this);
  }

  onChange(e) {
    super.onChange(e);
    this.props.onChange(this.props.index, this.state.object);
  }

  onDelete(e) {
    e.preventDefault();
    this.props.onDelete(this.props.index);
  }

  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    return(
      <Grid container spacing={4}>
        <Grid item xs={4}>
          <TextField
            id="key"
            label="Header name"
            margin="normal"
            value={this.state.object.key || ""}
            onChange={this.onChange}
            fullWidth
          />
        </Grid>
        <Grid item xs={7}>
          <TextField
            id="value"
            label="Header value"
            margin="normal"
            value={this.state.object.value || ""}
            onChange={this.onChange}
            fullWidth
          />
        </Grid>
        <Grid item xs={1} className={this.props.classes.delete}>
          <IconButton aria-label="delete" onClick={this.onDelete}>
            <Delete />
          </IconButton>
        </Grid>
      </Grid>
    );    
  }
}


HTTPIntegrationHeaderForm = withStyles(styles)(HTTPIntegrationHeaderForm);


class HTTPIntegrationForm extends FormComponent {
  constructor() {
    super();
    this.addHeader = this.addHeader.bind(this);
    this.onDeleteHeader = this.onDeleteHeader.bind(this);
    this.onChangeHeader = this.onChangeHeader.bind(this);
  }

  onChange(e) {
    super.onChange(e);
    this.props.onChange(this.state.object);
  }

  addHeader(e) {
    e.preventDefault();

    let object = this.state.object;
    if(object.headers === undefined) {
      object.headers = [{}];
    } else {
      object.headers.push({});
    }

    this.props.onChange(object);
  }

  onDeleteHeader(index) {
    let object = this.state.object;
    object.headers.splice(index, 1);
    this.props.onChange(object);
  }

  onChangeHeader(index, header) {
    let object = this.state.object;
    object.headers[index] = header;
    this.props.onChange(object);
  }

  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    let headers = [];
    if (this.state.object.headers !== undefined) {
      headers = this.state.object.headers.map((h, i) => <HTTPIntegrationHeaderForm key={i} index={i} object={h} onChange={this.onChangeHeader} onDelete={this.onDeleteHeader} />);
    }

    return(
      <div>
        <FormControl fullWidth margin="normal">
          <FormLabel>Headers</FormLabel>
          {headers}
        </FormControl>
        <Button variant="outlined" onClick={this.addHeader}>Add header</Button>
        <FormControl fullWidth margin="normal">
          <FormLabel>Endpoints</FormLabel>
          <TextField
            id="uplinkDataURL"
            label="Uplink data URL(s)"
            placeholder="http://example.com/uplink"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.uplinkDataURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />
          <TextField
            id="joinNotificationURL"
            label="Join notification URL(s)"
            placeholder="http://example.com/join"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.joinNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />
          <TextField
            id="statusNotificationURL"
            label="Device-status notification URL(s)"
            placeholder="http://example.com/status"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.statusNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />
          <TextField
            id="locationNotificationURL"
            label="Location notification URL(s)"
            placeholder="http://example.com/location"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.locationNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />
          <TextField
            id="ackNotificationURL"
            label="ACK notification URL(s)"
            placeholder="http://example.com/ack"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.ackNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />
          <TextField
            id="errorNotificationURL"
            label="Error notification URL(s)"
            placeholder="http://example.com/error"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.errorNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />
        </FormControl>
      </div>
    );
  }
}


class InfluxDBIntegrationForm extends FormComponent {
  onChange(e) {
    super.onChange(e);
    this.props.onChange(this.state.object);
  }

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
      return(<div></div>);
    }

    return(
      <FormControl fullWidth margin="normal">
        <FormLabel>InfluxDB integration configuration</FormLabel>
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
      </FormControl>
    );
  }
}

InfluxDBIntegrationForm = withStyles(styles)(InfluxDBIntegrationForm);


class ThingsBoardIntegrationForm extends FormComponent {
  onChange(e) {
    super.onChange(e);
    this.props.onChange(this.state.object);
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <FormControl fullWidth margin="normal">
        <FormLabel>ThingsBoard.io integration configuration</FormLabel>
        <TextField
          id="server"
          label="ThingsBoard.io server"
          placeholder="http://host:port"
          value={this.state.object.server || ""}
          onChange={this.onChange}
          margin="normal"
          required
          fullWidth
        />
        <FormHelperText>
          Each device must have a 'ThingsBoardAccessToken' variable assigned. This access-token is generated by ThingsBoard.
        </FormHelperText>
      </FormControl>
    );
  }
}

ThingsBoardIntegrationForm = withStyles(styles)(ThingsBoardIntegrationForm);


class IntegrationForm extends FormComponent {
  constructor() {
    super();
    this.getKindOptions = this.getKindOptions.bind(this);
    this.onFormChange = this.onFormChange.bind(this);
  }

  onFormChange(object) {
    this.setState({
      object: object,
    });
  }

  getKindOptions(search, callbackFunc) {
    const kindOptions = [
      {value: "http", label: "HTTP integration"},
      {value: "influxdb", label: "InfluxDB integration"},
      {value: "thingsboard", label: "ThingsBoard.io"},
    ];

    callbackFunc(kindOptions);
  }

  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        {!this.props.update && <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Integration kind</FormLabel>
          <AutocompleteSelect
            id="kind"
            label="Select integration kind"
            value={this.state.object.kind || ""}
            onChange={this.onChange}
            getOptions={this.getKindOptions}
          />
        </FormControl>}
        {this.state.object.kind === "http" && <HTTPIntegrationForm object={this.state.object} onChange={this.onFormChange} />}
        {this.state.object.kind === "influxdb" && <InfluxDBIntegrationForm object={this.state.object} onChange={this.onFormChange} />}
        {this.state.object.kind === "thingsboard" && <ThingsBoardIntegrationForm object={this.state.object} onChange={this.onFormChange} />}
      </Form>
    );
  }
}

export default withStyles(styles)(IntegrationForm);
