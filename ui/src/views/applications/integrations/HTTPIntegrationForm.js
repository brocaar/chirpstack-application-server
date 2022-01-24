import React from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import TextField from '@material-ui/core/TextField';
import FormHelperText from "@material-ui/core/FormHelperText";
import IconButton from '@material-ui/core/IconButton';
import Button from "@material-ui/core/Button";

import Delete from "mdi-material-ui/Delete";

import FormComponent from "../../../classes/FormComponent";
import Form from "../../../components/Form";
import AutocompleteSelect from "../../../components/AutocompleteSelect";
import theme from "../../../theme";


const styles = {
  delete: {
    marginTop: 3 * theme.spacing(1),
  },
  formLabel: {
    fontSize: 12,
  },
};


class HTTPIntegrationHeaderForm extends FormComponent {
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
  addHeader = (e) => {
    e.preventDefault();

    let object = this.state.object;
    if(object.headers === undefined) {
      object.headers = [{}];
    } else {
      object.headers.push({});
    }

    this.setState({
      object: object,
    });
  }

  onDeleteHeader = (index) => {
    let object = this.state.object;
    object.headers.splice(index, 1);

    this.setState({
      object: object,
    });
  }

  onChangeHeader = (index, header) => {
    let object = this.state.object;
    object.headers[index] = header;
    this.setState({
      object: object,
    });
  }

  getMarshalerOptions = (search, callbackFunc) => {
    const marshalerOptions = [
      {value: "JSON", label: "JSON"},
      {value: "PROTOBUF", label: "Protocol Buffers"},
      {value: "JSON_V3", label: "JSON (legacy, will be deprecated)"},
    ];

    callbackFunc(marshalerOptions);
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    let headers = [];
    if (this.state.object.headers !== undefined) {
      headers = this.state.object.headers.map((h, i) => <HTTPIntegrationHeaderForm key={i} index={i} object={h} onChange={this.onChangeHeader} onDelete={this.onDeleteHeader} />);
    }

    return(
      <Form submitLabel={this.props.submitLabel} onSubmit={this.onSubmit}>
        <FormControl fullWidth margin="normal">
          <FormLabel required>Payload marshaler</FormLabel>
          <AutocompleteSelect
            id="marshaler"
            label="Select payload marshaler"
            value={this.state.object.marshaler || ""}
            onChange={this.onChange}
            getOptions={this.getMarshalerOptions}
          />
          <FormHelperText>This defines how the payload will be encoded.</FormHelperText>
        </FormControl>
        <FormControl fullWidth margin="normal">
          <FormLabel>Headers</FormLabel>
          {headers}
        </FormControl>
        <Button variant="outlined" onClick={this.addHeader}>Add header</Button>
        <FormControl fullWidth margin="normal">
          <FormLabel>Endpoints</FormLabel>
          <TextField
            id="eventEndpointURL"
            label="Endpoint URL(s) for events"
            placeholder="http://example.com/events"
            helperText="ChirpStack will make a POST request to this URL(s) with 'event' as query parameter. Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.eventEndpointURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />
          {!!this.state.object.uplinkDataURL && <TextField
            id="uplinkDataURL"
            label="Uplink data URL(s)"
            placeholder="http://example.com/uplink"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.uplinkDataURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
          {!!this.state.object.joinNotificationURL && <TextField
            id="joinNotificationURL"
            label="Join notification URL(s)"
            placeholder="http://example.com/join"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.joinNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
          {!!this.state.object.statusNotificationURL && <TextField
            id="statusNotificationURL"
            label="Device-status notification URL(s)"
            placeholder="http://example.com/status"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.statusNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
          {!!this.state.object.locationNotificationURL && <TextField
            id="locationNotificationURL"
            label="Location notification URL(s)"
            placeholder="http://example.com/location"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.locationNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
          {!!this.state.object.ackNotificationURL && <TextField
            id="ackNotificationURL"
            label="ACK notification URL(s)"
            placeholder="http://example.com/ack"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.ackNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
          {!!this.state.object.txAckNotificationURL && <TextField
            id="txAckNotificationURL"
            label="TX ACK notification URL(s)"
            placeholder="http://example.com/txack"
            helperText="This notification is sent when the downlink was acknowledged by the LoRa gateway for transmission. Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.txAckNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
          {!!this.state.object.integrationNotificationURL && <TextField
            id="integrationNotificationURL"
            label="Integration notification URL(s)"
            placeholder="http://example.com/integration"
            helperText="This notification can by sent by configured integrations to send custom payloads. Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.integrationNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
          {!!this.state.object.errorNotificationURL && <TextField
            id="errorNotificationURL"
            label="Error notification URL(s)"
            placeholder="http://example.com/error"
            helperText="Multiple URLs can be defined as a comma separated list. Whitespace will be automatically removed."
            value={this.state.object.errorNotificationURL || ""}
            onChange={this.onChange}
            margin="normal"
            fullWidth
          />}
        </FormControl>
      </Form>
    );
  }
}


export default HTTPIntegrationForm;
