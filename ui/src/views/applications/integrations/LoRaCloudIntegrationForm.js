import React from "react";

import FormControl from "@material-ui/core/FormControl";
import TextField from '@material-ui/core/TextField';
import Tabs from "@material-ui/core/Tabs";
import Tab from "@material-ui/core/Tab";
import FormGroup from "@material-ui/core/FormGroup";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import FormHelperText from "@material-ui/core/FormHelperText";
import Checkbox from "@material-ui/core/Checkbox";
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import ChevronDown from "mdi-material-ui/ChevronDown";

import FormComponent from "../../../classes/FormComponent";
import Form from "../../../components/Form";


class LoRaCloudIntegrationForm extends FormComponent {
  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <Form submitLabel={this.props.submitLabel} onSubmit={this.onSubmit}>
        <Tabs 
          value={0}
          indicatorColor="primary"
        >
          <Tab label="Modem & Geolocation Services" />
        </Tabs>
        <TextField
          id="dasToken"
          label="Token"
          value={this.state.object.dasToken || ""}
          onChange={this.onChange}
          margin="normal"
          helperText="This token can be obtained from loracloud.com"
          type="password"
          required
          fullWidth
        />
        <FormControl fullWidth margin="normal">
          <FormGroup>
            <FormControlLabel
              label="I am using LoRa Edge&trade; LR1110 or my device uses LoRa Basics&trade; Modem-E"
              control={
                <Checkbox 
                  id="das"
                  checked={!!this.state.object.das}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
          </FormGroup>
        </FormControl>
        {!!this.state.object.das && <TextField
          id="dasGNSSPort"
          label="GNSS port (FPort)"
          value={this.state.object.dasGNSSPort || 0}
          onChange={this.onChange}
          type="number"
          margin="normal"
          helperText="ChirpStack Application Server will only forward the FRMPayload for GNSS geolocation to LoRa Cloud when the uplink matches the configured port."
          fullWidth
        />}
        {!!this.state.object.das && <TextField
          id="dasModemPort"
          label="Modem port (FPort)"
          value={this.state.object.dasModemPort || 0}
          onChange={this.onChange}
          type="number"
          margin="normal"
          helperText="ChirpStack Application Server will only forward the FRMPayload to LoRa Cloud when the uplink matches the configured port."
          fullWidth
        />}
        {!!this.state.object.das && <FormControl fullWidth margin="normal">
          <FormGroup>
            <FormControlLabel
              label="Use receive timestamp for GNSS geolocation"
              control={
                <Checkbox 
                  id="dasGNSSUseRxTime"
                  checked={!!this.state.object.dasGNSSUseRxTime}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
            <FormHelperText>
              If enabled, the receive timestamp of the gateway will be used as reference instead of the timestamp
              included in the GNSS payload.
            </FormHelperText>
          </FormGroup>
        </FormControl>}
        {!!this.state.object.das && <FormControl fullWidth margin="normal">
          <FormGroup>
            <FormControlLabel
              label="My device adheres to the LoRa Edge&trade; Tracker Reference Design protocol"
              control={
                <Checkbox 
                  id="dasStreamingGeolocWorkaround"
                  checked={!!this.state.object.dasStreamingGeolocWorkaround}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
            <FormHelperText>
              If enabled, ChirpStack Application Server will try to resolve the location of the device if a geolocation payload is detected.
            </FormHelperText>
          </FormGroup>
        </FormControl>}
        <br /><br />
        <ExpansionPanel>
          <ExpansionPanelSummary expandIcon={<ChevronDown />}>Advanced geolocation options</ExpansionPanelSummary>
          <ExpansionPanelDetails>
            <div>
            <FormControl fullWidth margin="normal">
              <FormGroup>
                <FormControlLabel
                  label="TDOA based geolocation"
                  control={
                    <Checkbox 
                      id="geolocationTDOA"
                      checked={!!this.state.object.geolocationTDOA}
                      onChange={this.onChange}
                      color="primary"
                    />
                  }
                />
              </FormGroup>
              <FormHelperText>
                If enabled, geolocation will be based on time-difference of arrival (TDOA). Please note that this
                requires gateways that support the fine-timestamp feature.
              </FormHelperText>
            </FormControl>
            <FormControl fullWidth margin="normal">
              <FormGroup>
                <FormControlLabel
                  label="RSSI based geolocation"
                  control={
                    <Checkbox 
                      id="geolocationRSSI"
                      checked={!!this.state.object.geolocationRSSI}
                      onChange={this.onChange}
                      color="primary"
                    />
                  }
                />
              </FormGroup>
              <FormHelperText>
                If enabled, geolocation will be based on RSSI values reported by the receiving gateways.
              </FormHelperText>
            </FormControl>
            <FormControl fullWidth margin="normal">
              <FormGroup>
                <FormControlLabel
                  label="Wi-Fi-based geolocation"
                  control={
                    <Checkbox 
                      id="geolocationWifi"
                      checked={!!this.state.object.geolocationWifi}
                      onChange={this.onChange}
                      color="primary"
                    />
                  }
                />
              </FormGroup>
              <FormHelperText>
                If enabled, geolocation will be based on Wi-Fi access-point data reported by the device.
              </FormHelperText>
            </FormControl>
            <FormControl fullWidth margin="normal">
              <FormGroup>
                <FormControlLabel
                  label="GNSS-based geolocation (LR1110)"
                  control={
                    <Checkbox 
                      id="geolocationGNSS"
                      checked={!!this.state.object.geolocationGNSS}
                      onChange={this.onChange}
                      color="primary"
                    />
                  }
                />
              </FormGroup>
              <FormHelperText>
                If enabled, geolocation will be based on GNSS data reported by the device.
              </FormHelperText>
            </FormControl>
            {(this.state.object.geolocationTDOA || this.state.object.geolocationRSSI) && <TextField
              id="geolocationBufferTTL"
              label="Geolocation buffer TTL (seconds)"
              type="number"
              margin="normal"
              value={this.state.object.geolocationBufferTTL || 0}
              onChange={this.onChange}
              helperText="The time in seconds that historical uplinks will be stored in the geolocation buffer. Used for TDOA and RSSI geolocation."
              fullWidth
            />}
            {(this.state.object.geolocationTDOA || this.state.object.geolocationRSSI) && <TextField
              id="geolocationMinBufferSize"
              label="Geolocation minimum buffer size"
              type="number"
              margin="normal"
              value={this.state.object.geolocationMinBufferSize || 0}
              onChange={this.onChange}
              helperText="The minimum buffer size required before using geolocation. Using multiple uplinks for geolocation can increase the accuracy of the geolocation results. Used for TDOA and RSSI geolocation."
              fullWidth
            />}
            {this.state.object.geolocationWifi && <TextField
              id="geolocationWifiPayloadField"
              label="Wifi payload field"
              value={this.state.object.geolocationWifiPayloadField || ""}
              onChange={this.onChange}
              margin="normal"
              helperText="This must match the name of the field in the decoded payload which holds array of Wifi access-points. Each element in the array must contain two keys: 1) macAddress: array of 6 bytes, 2) signalStrength: RSSI of the access-point."
              required
              fullWidth
            />}
            {this.state.object.geolocationGNSS && <TextField
              id="geolocationGNSSPayloadField"
              label="GNSS payload field"
              value={this.state.object.geolocationGNSSPayloadField || ""}
              onChange={this.onChange}
              margin="normal"
              helperText="This must match the name of the field in the decoded payload which holds the LR1110 GNSS bytes."
              required
              fullWidth
            />}
            {this.state.object.geolocationGNSS && <FormControl fullWidth margin="normal">
              <FormGroup>
                <FormControlLabel
                  label="Use receive timestamp for GNSS geolocation"
                  control={
                    <Checkbox 
                      id="geolocationGNSSUseRxTime"
                      checked={!!this.state.object.geolocationGNSSUseRxTime}
                      onChange={this.onChange}
                      color="primary"
                    />
                  }
                />
              </FormGroup>
              <FormHelperText>
                If enabled, the receive timestamp of the gateway will be used as reference instead of the timestamp included in the GNSS payload.
              </FormHelperText>
            </FormControl>}
            </div>
          </ExpansionPanelDetails>
        </ExpansionPanel>
      </Form>
    );
  }
}


export default LoRaCloudIntegrationForm;
