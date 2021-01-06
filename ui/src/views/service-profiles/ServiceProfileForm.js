import React from "react";

import TextField from '@material-ui/core/TextField';
import FormLabel from "@material-ui/core/FormLabel";
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormGroup from "@material-ui/core/FormGroup";
import Checkbox from '@material-ui/core/Checkbox';
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import AutocompleteSelect from "../../components/AutocompleteSelect";
import NetworkServerStore from "../../stores/NetworkServerStore";


class ServiceProfileForm extends FormComponent {
  constructor() {
    super();
    this.getNetworkServerOption = this.getNetworkServerOption.bind(this);
    this.getNetworkServerOptions = this.getNetworkServerOptions.bind(this);
  }

  getNetworkServerOption(id, callbackFunc) {
    NetworkServerStore.get(id, resp => {
      callbackFunc({label: resp.networkServer.name, value: resp.networkServer.id});
    });
  }

  getNetworkServerOptions(search, callbackFunc) {
    NetworkServerStore.list(0, 999, 0, resp => {
      const options = resp.result.map((ns, i) => {return {label: ns.name, value: ns.id}});
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
          label="Service-profile name"
          margin="normal"
          value={this.state.object.name || ""}
          onChange={this.onChange}
          helperText="A name to identify the service-profile."
          disabled={this.props.disabled}
          required
          fullWidth
        />
        {!this.props.update && <FormControl fullWidth margin="normal">
          <FormLabel required>Network-server</FormLabel>
          <AutocompleteSelect
            id="networkServerID"
            label="Network-server"
            value={this.state.object.networkServerID || null}
            onChange={this.onChange}
            getOption={this.getNetworkServerOption}
            getOptions={this.getNetworkServerOptions}
          />
          <FormHelperText>
            The network-server on which this service-profile will be provisioned. After creating the service-profile, this value can't be changed.
          </FormHelperText>
        </FormControl>}
        <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="Add gateway meta-data"
            control={
              <Checkbox
                id="addGWMetaData"
                checked={!!this.state.object.addGWMetaData}
                onChange={this.onChange}
                disabled={this.props.disabled}
                color="primary"
              />
            }
          />
          <FormHelperText>
            GW metadata (RSSI, SNR, GW geoloc., etc.) are added to the packet sent to the application-server.
          </FormHelperText>
        </FormControl>
        <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="Enable network geolocation"
            control={
              <Checkbox
                id="nwkGeoLoc"
                checked={!!this.state.object.nwkGeoLoc}
                onChange={this.onChange}
                disabled={this.props.disabled}
                color="primary"
              />
            }
          />
          <FormHelperText>
            When enabled, the network-server will try to resolve the location of the devices under this service-profile.
            Please note that you need to have gateways supporting the fine-timestamp feature and that the network-server
            needs to be configured in order to provide geolocation support.
          </FormHelperText>
        </FormControl>
        <TextField
          id="devStatusReqFreq"
          label="Device-status request frequency"
          margin="normal"
          type="number"
          value={this.state.object.devStatusReqFreq || 0}
          onChange={this.onChange}
          helperText="Frequency to initiate an End-Device status request (request/day). Set to 0 to disable."
          disabled={this.props.disabled}
          fullWidth
        />
        {this.state.object.devStatusReqFreq > 0 && <FormControl fullWidth margin="normal">
          <FormGroup>
            <FormControlLabel
              label="Report device battery level to application-server"
              control={
                <Checkbox
                  id="reportDevStatusBattery"
                  checked={!!this.state.object.reportDevStatusBattery}
                  onChange={this.onChange}
                  disabled={this.props.disabled}
                  color="primary"
                />
              }
            />
            <FormControlLabel
              label="Report device link margin to application-server"
              control={
                <Checkbox
                  id="reportDevStatusMargin"
                  checked={!!this.state.object.reportDevStatusMargin}
                  onChange={this.onChange}
                  disabled={this.props.disabled}
                  color="primary"
                />
              }
            />
          </FormGroup>
        </FormControl>}
        <TextField
          id="drMin"
          label="Minimum allowed data-rate"
          margin="normal"
          type="number"
          value={this.state.object.drMin || 0}
          onChange={this.onChange}
          helperText="Minimum allowed data rate. Used for ADR."
          disabled={this.props.disabled}
          fullWidth
          required
        />
        <TextField
          id="drMax"
          label="Maximum allowed data-rate"
          margin="normal"
          type="number"
          value={this.state.object.drMax || 0}
          onChange={this.onChange}
          helperText="Maximum allowed data rate. Used for ADR."
          disabled={this.props.disabled}
          fullWidth
          required
        />
        <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="Private gateways"
            control={
              <Checkbox
                id="gwsPrivate"
                checked={!!this.state.object.gwsPrivate}
                onChange={this.onChange}
                disabled={this.props.disabled}
                color="primary"
              />
            }
          />
          <FormHelperText>
            Gateways under this service-profile are private. This means that these gateways can only be used by devices under the same service-profile.
          </FormHelperText>
        </FormControl>
      </Form>
    );
  }
}

export default ServiceProfileForm;
