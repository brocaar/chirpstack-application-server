import React from "react";

import { withStyles } from "@material-ui/core/styles";
import TextField from '@material-ui/core/TextField';
import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import FormHelperText from "@material-ui/core/FormHelperText";
import Button from "@material-ui/core/Button";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import DurationField from "../../components/DurationField";
import AutocompleteSelect from "../../components/AutocompleteSelect";

const styles = {
  formLabel: {
    fontSize: 12,
  },
};

class FUOTADeploymentForm extends FormComponent {
  constructor() {
    super();

    this.state.file = null;

    this.onFileChange = this.onFileChange.bind(this);
  }

  getGroupTypeOptions(search, callbackFunc) {
    const options = [
      {value: "CLASS_C", label: "Class-C"},
    ];

    callbackFunc(options);
  }

  getMulticastTimeoutOptions(search, callbackFunc) {
    let options = [];

    for (let i = 0; i < (1 << 4); i++) {
      options.push({
        label: `${1 << i} seconds`,
        value: i,
      });
    }

    callbackFunc(options);
  }

  onFileChange(e) {
    let object = this.state.object;

    if (e.target.files.length !== 1) {
      object.payload = "";

      this.setState({
        file: null,
        object: object,
      });
    } else {
      this.setState({
        file: e.target.files[0],
      });

      const reader = new FileReader();
      reader.onload = () => {
        const encoded = reader.result.replace(/^data:(.*;base64,)?/, '');
        object.payload = encoded;

        this.setState({
          object: object,
        });
      };
      reader.readAsDataURL(e.target.files[0]);
    }
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    let fileLabel = "";
    if (this.state.file !== null) {
      fileLabel = `${this.state.file.name} (${this.state.file.size} bytes)`
    } else {
      fileLabel = "Select file..."
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <TextField
          id="name"
          label="Firmware update job-name"
          helperText="A descriptive name for this firmware update job."
          margin="normal"
          value={this.state.object.name || ""}
          onChange={this.onChange}
          fullWidth
          required
        />

        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Select firmware file</FormLabel>
          <Button component="label">
            {fileLabel}
            <input type="file" style={{display: "none"}} onChange={this.onFileChange} />
          </Button>
          <FormHelperText>
            This file will fragmented and sent to the device(s). Please note that the format of this file is vendor dependent.
          </FormHelperText>
        </FormControl>

        <TextField
          id="redundancy"
          label="Redundant frames"
          helperText="The given number represents the extra redundant frames that will be sent so that a device can recover from packet-loss."
          margin="normal"
          type="number"
          value={this.state.object.redundancy || 0}
          onChange={this.onChange}
          required
          fullWidth
        />

        <DurationField
          id="unicastTimeout"
          label="Unicast timeout (seconds)"
          helperText="Set this to the minimum interval in which the device(s) are sending uplink messages."
          value={this.state.object.unicastTimeout}
          onChange={this.onChange}
        />

        <TextField
          id="dr"
          label="Data-rate"
          helperText="The data-rate to use when transmitting the multicast frames. Please refer to the LoRaWAN Regional Parameters specification for valid values."
          margin="normal"
          type="number"
          value={this.state.object.dr || 0}
          onChange={this.onChange}
          required
          fullWidth
        />

        <TextField
          id="frequency"
          label="Frequency (Hz)"
          helperText="The frequency to use when transmitting the multicast frames. Please refer to the LoRaWAN Regional Parameters specification for valid values."
          margin="normal"
          type="number"
          value={this.state.object.frequency || 0}
          onChange={this.onChange}
          required
          fullWidth
        />

        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Multicast-group type</FormLabel>
          <AutocompleteSelect
            id="groupType"
            label="Select multicast-group type"
            value={this.state.object.groupType || ""}
            onChange={this.onChange}
            getOptions={this.getGroupTypeOptions}
          />
          <FormHelperText>
            The multicast-group type defines the way how multicast frames are scheduled by the network-server.
          </FormHelperText>
        </FormControl>

        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Multicast timeout</FormLabel>
          <AutocompleteSelect
            id="multicastTimeout"
            label="Select multicast timeout"
            value={this.state.object.multicastTimeout || ""}
            onChange={this.onChange}
            getOptions={this.getMulticastTimeoutOptions}
          />
        </FormControl>

      </Form>
    );
  }
}

export default withStyles(styles)(FUOTADeploymentForm);

