import React from "react";

import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import TextField from '@material-ui/core/TextField';
import FormHelperText from "@material-ui/core/FormHelperText";

import FormComponent from "../../../classes/FormComponent";
import Form from "../../../components/Form";
import AutocompleteSelect from "../../../components/AutocompleteSelect";


class GCPPubSubIntegrationForm extends FormComponent {
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
            required
          />
          <FormHelperText>This defines how the payload will be encoded.</FormHelperText>
        </FormControl>
        <TextField
          id="projectID"
          label="GCP project ID"
          value={this.state.object.projectID || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />
        <TextField
          id="topicName"
          label="GCP Pub/Sub topic name"
          value={this.state.object.topicName || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />
        <TextField
          id="credentialsFile"
          label="GCP Service account credentials file"
          value={this.state.object.credentialsFile || ""}
          onChange={this.onChange}
          margin="normal"
          rows={10}
          helperText="Under IAM create a Service account with 'Pub/Sub Publisher' role, then put the content of the JSON key in this field."
          fullWidth
          multiline
          required
        />
      </Form>
    );
  }
}


export default GCPPubSubIntegrationForm;
