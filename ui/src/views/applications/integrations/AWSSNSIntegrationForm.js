import React from "react";

import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import TextField from '@material-ui/core/TextField';
import FormHelperText from "@material-ui/core/FormHelperText";

import FormComponent from "../../../classes/FormComponent";
import Form from "../../../components/Form";
import AutocompleteSelect from "../../../components/AutocompleteSelect";


class AWSSNSIntegrationForm extends FormComponent {
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
          id="region"
          label="AWS region"
          value={this.state.object.region || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />
        <TextField
          id="accessKeyID"
          label="AWS Access Key ID"
          value={this.state.object.accessKeyID || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />
        <TextField
          id="secretAccessKey"
          label="AWS Secret Access Key"
          value={this.state.object.secretAccessKey || ""}
          onChange={this.onChange}
          margin="normal"
          type="password"
          fullWidth
          required
        />
        <TextField
          id="topicARN"
          label="AWS SNS topic ARN"
          value={this.state.object.topicARN || ""}
          onChange={this.onChange}
          margin="normal"
          fullWidth
          required
        />
      </Form>
    );
  }
}


export default AWSSNSIntegrationForm;
