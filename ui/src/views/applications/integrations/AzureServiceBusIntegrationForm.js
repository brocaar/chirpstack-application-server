import React from "react";

import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import TextField from '@material-ui/core/TextField';
import FormHelperText from "@material-ui/core/FormHelperText";

import FormComponent from "../../../classes/FormComponent";
import Form from "../../../components/Form";
import AutocompleteSelect from "../../../components/AutocompleteSelect";


class AzureServiceBusIntegrationForm extends FormComponent {
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
          id="connectionString"
          label="Azure Service-Bus connection string"
          value={this.state.object.connectionString || ""}
          onChange={this.onChange}
          margin="normal"
          helperText="This string can be obtained after creating a 'Shared access policy' with 'Send' permission."
          fullWidth
          required
        />
        <TextField
          id="publishName"
          label="Azure Service-Bus topic / queue name"
          value={this.state.object.publishName || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />
      </Form>
    );
  }
}

export default AzureServiceBusIntegrationForm;
