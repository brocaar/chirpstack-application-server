import React from "react";

import { withStyles } from "@material-ui/core/styles";
import TextField from '@material-ui/core/TextField';
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import FormLabel from "@material-ui/core/FormLabel";
import FormHelperText from "@material-ui/core/FormHelperText";
import Checkbox from "@material-ui/core/Checkbox";
import FormGroup from "@material-ui/core/FormGroup";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import AutocompleteSelect from "../../components/AutocompleteSelect";
import DeviceProfileStore from "../../stores/DeviceProfileStore";


const styles = {
  formLabel: {
    fontSize: 12,
  },
};


class DeviceForm extends FormComponent {
  constructor() {
    super();
    this.getDeviceProfileOption = this.getDeviceProfileOption.bind(this);
    this.getDeviceProfileOptions = this.getDeviceProfileOptions.bind(this);
  }

  getDeviceProfileOption(id, callbackFunc) {
    DeviceProfileStore.get(id, resp => {
      callbackFunc({label: resp.deviceProfile.name, value: resp.deviceProfile.id});
    });
  }

  getDeviceProfileOptions(search, callbackFunc) {
    DeviceProfileStore.list(0, this.props.match.params.applicationID, 999, 0, resp => {
      const options = resp.result.map((dp, i) => {return {label: dp.name, value: dp.id}});
      callbackFunc(options);
    });
  }

  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
        disabled={this.props.disabled}
      >
        <TextField
          id="name"
          label="Device name"
          helperText="The name may only contain words, numbers and dashes."
          margin="normal"
          value={this.state.object.name || ""}
          onChange={this.onChange}
          inputProps={{
            pattern: "[\\w-]+",
          }}
          fullWidth
          required
        />
        <TextField
          id="description"
          label="Device description"
          margin="normal"
          value={this.state.object.description || ""}
          onChange={this.onChange}
          fullWidth
          required
        />
        {!this.props.update && <TextField
          id="devEUI"
          label="Device EUI"
          placeholder="0000000000000000"
          helperText="The device EUI in hex encoding."
          onChange={this.onChange}
          value={this.state.object.devEUI || ""}
          inputProps={{
            pattern: "[A-Fa-f0-9]{16}"
          }}
          fullWidth
          required
        />}
        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Device-profile</FormLabel>
          <AutocompleteSelect
            id="deviceProfileID"
            label="Device-profile"
            value={this.state.object.deviceProfileID}
            onChange={this.onChange}
            getOption={this.getDeviceProfileOption}
            getOptions={this.getDeviceProfileOptions}
          />
        </FormControl>
        <FormControl margin="normal">
          <FormGroup>
            <FormControlLabel
              label="Disable frame-counter validation"
              control={
                <Checkbox
                  id="skipFCntCheck"
                  checked={!!this.state.object.skipFCntCheck}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
          </FormGroup>
          <FormHelperText>
            Note that disabling the frame-counter validation will compromise security as it enables people to perform replay-attacks.
          </FormHelperText>
        </FormControl>
      </Form>
    );
  }
}

export default withStyles(styles)(DeviceForm);
