import React from "react";

import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import TextField from '@material-ui/core/TextField';

import FormComponent from "../../../classes/FormComponent";
import Form from "../../../components/Form";
import AutocompleteSelect from "../../../components/AutocompleteSelect";


class MyDevicesIntegrationForm extends FormComponent {
  getEndpointOptions(search, callbackFunc) {
    const endpointOptions = [
      {value: "https://lora.mydevices.com/v1/networks/chirpstackio/uplink", label: "Cayenne"},
      {value: "https://lora.iotinabox.com/v1/networks/iotinabox.chirpstackio/uplink", label: "IoT in a Box"},
      {value: "custom", label: "Custom endpoint URL"},
    ];

    callbackFunc(endpointOptions);
  }

  endpointChange = (e) => {
    let object = this.state.object;

    if (e.target.value === "custom") {
      object.endpoint = "";
    } else {
      object.endpoint = e.target.value;
    }

    this.setState({
      object: object,
    });
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    let endpointSelect = "custom";
    if (this.state.object.endpoint === undefined) {
      endpointSelect = "";
    }

    this.getEndpointOptions("", (options) => {
      for (let opt of options) {
        if (this.state.object.endpoint === opt.value) {
          endpointSelect = this.state.object.endpoint;
        }
      }
    });

    return(
      <Form submitLabel={this.props.submitLabel} onSubmit={this.onSubmit}>
        <FormControl fullWidth margin="normal">
          <FormLabel>myDevices endpoint</FormLabel>
          <AutocompleteSelect
            id="_endpoint"
            label="Select myDevices endpoint"
            value={endpointSelect || ""}
            getOptions={this.getEndpointOptions}
            onChange={this.endpointChange}
          />
        </FormControl>
        {endpointSelect === "custom" && <FormControl fullWidth margin="normal">
          <FormLabel>myDevices integration configuration</FormLabel>
          <TextField
            id="endpoint"
            label="myDevices API endpoint"
            placeholder="http://host:port"
            value={this.state.object.endpoint || ""}
            onChange={this.onChange}
            margin="normal"
            required
            fullWidth
          />
        </FormControl>}
      </Form>
    );
  }
}


export default MyDevicesIntegrationForm;
