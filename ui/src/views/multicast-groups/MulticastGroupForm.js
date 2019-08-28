import React from "react";

import { withStyles } from "@material-ui/core/styles";
import TextField from '@material-ui/core/TextField';
import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import FormHelperText from "@material-ui/core/FormHelperText";

import FormComponent from "../../classes/FormComponent";
import AESKeyField from "../../components/AESKeyField";
import DevAddrField from "../../components/DevAddrField";
import Form from "../../components/Form";
import AutocompleteSelect from "../../components/AutocompleteSelect";
import ServiceProfileStore from "../../stores/ServiceProfileStore";
import theme from "../../theme";


const styles = {
  formLabel: {
    fontSize: 12,
  },
  link: {
    color: theme.palette.primary.main,
  },
};


class MulticastGroupForm extends FormComponent {
  constructor() {
    super();
    this.getServiceProfileOption = this.getServiceProfileOption.bind(this);
    this.getServiceProfileOptions = this.getServiceProfileOptions.bind(this);
  }

  getServiceProfileOption(id, callbackFunc) {
    ServiceProfileStore.get(id, resp => {
      callbackFunc({label: resp.serviceProfile.name, value: resp.serviceProfile.id});
    });
  }

  getServiceProfileOptions(search, callbackFunc) {
    ServiceProfileStore.list(this.props.match.params.organizationID, 999, 0, resp => {
      const options = resp.result.map((sp, i) => {return {label: sp.name, value: sp.id}});
      callbackFunc(options);
    });
  }

  getRandomKey(len) {
    let key = "";
    const possible = 'abcdef0123456789';

    for(let i = 0; i < len; i++){
      key += possible.charAt(Math.floor(Math.random() * possible.length));
    }

    return key;
  }

  getRandomMcAddr = (cb) => {
    cb(this.getRandomKey(8));
  }

  getRandomSessionKey = (cb) => {
    cb(this.getRandomKey(32));
  }


  getGroupTypeOptions(search, callbackFunc) {
    const options = [
      {value: "CLASS_B", label: "Class-B"},
      {value: "CLASS_C", label: "Class-C"},
    ];

    callbackFunc(options);
  }

  getPingSlotPeriodOptions(search, callbackFunc) {
    const pingSlotPeriodOptions = [
      {value: 32 * 1, label: "every second"},
      {value: 32 * 2, label: "every 2 seconds"},
      {value: 32 * 4, label: "every 4 seconds"},
      {value: 32 * 8, label: "every 8 seconds"},
      {value: 32 * 16, label: "every 16 seconds"},
      {value: 32 * 32, label: "every 32 seconds"},
      {value: 32 * 64, label: "every 64 seconds"},
      {value: 32 * 128, label: "every 128 seconds"},
    ];

    callbackFunc(pingSlotPeriodOptions);
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <TextField
          id="name"
          label="Multicast-group name"
          margin="normal"
          value={this.state.object.name || ""}
          onChange={this.onChange}
          helperText="The name of the multicast-group."
          fullWidth
          required
        />
        {!this.props.update && <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Service-profile</FormLabel> 
          <AutocompleteSelect
            id="serviceProfileID"
            label="Select service-profile"
            value={this.state.object.serviceProfileID || ""}
            onChange={this.onChange}
            getOption={this.getServiceProfileOption}
            getOptions={this.getServiceProfileOptions}
            margin="none"
          />
          <FormHelperText>
            The service-profile to which this multicast-group will be attached. Note that you can't change this value after the multicast-group has been created.
          </FormHelperText>
        </FormControl>}
        <DevAddrField
          id="mcAddr"
          label="Multicast address"
          margin="normal"
          value={this.state.object.mcAddr || ""}
          onChange={this.onChange}
          disabled={this.props.disabled}
          randomFunc={this.getRandomMcAddr}
          fullWidth
          required
          random
        />
        <AESKeyField
          id="mcNwkSKey"
          label="Multicast network session key"
          margin="normal"
          value={this.state.object.mcNwkSKey || ""}
          onChange={this.onChange}
          disabled={this.props.disabled}
          fullWidth
          required
          random
        />
        <AESKeyField
          id="mcAppSKey"
          label="Multicast application session key"
          margin="normal"
          value={this.state.object.mcAppSKey || ""}
          onChange={this.onChange}
          disabled={this.props.disabled}
          fullWidth
          required
          random
        />
        <TextField
          id="fCnt"
          label="Frame-counter"
          margin="normal"
          type="number"
          value={this.state.object.fCnt || 0}
          onChange={this.onChange}
          required
          fullWidth
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
        {this.state.object.groupType === "CLASS_B" && <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Class-B ping-slot periodicity</FormLabel>
          <AutocompleteSelect
            id="pingSlotPeriod"
            label="Select Class-B ping-slot periodicity"
            value={this.state.object.pingSlotPeriod || ""}
            onChange={this.onChange}
            getOptions={this.getPingSlotPeriodOptions}
          />
          <FormHelperText>Class-B ping-slot periodicity.</FormHelperText>
        </FormControl>}
      </Form>
    );
  }
}

export default withStyles(styles)(MulticastGroupForm);
