import React from "react";

import TextField from '@material-ui/core/TextField';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';

import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormGroup from "@material-ui/core/FormGroup";
import FormHelperText from '@material-ui/core/FormHelperText';
import Checkbox from '@material-ui/core/Checkbox';

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import FormControl from "../../components/FormControl";


class NetworkServerForm extends FormComponent {
  constructor() {
    super();
    this.state = {
      tab: 0,
    };

    this.onChangeTab = this.onChangeTab.bind(this);
  }

  onChangeTab(event, value) {
    this.setState({
      tab: value,
    });
  }

  render() {
    if (this.state.object === undefined) {
      return(null);
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
            <Tabs
              value={this.state.tab}
              indicatorColor="primary"
              textColor="primary"
              onChange={this.onChangeTab}
            >
              <Tab label="General" />
              <Tab label="Gateway discovery" />
              <Tab label="TLS certificates" />
            </Tabs>
          {this.state.tab === 0 && <div>
            <TextField
              id="name"
              label="Network-server name"
              fullWidth={true}
              margin="normal"
              value={this.state.object.name || ""}
              onChange={this.onChange}
              helperText="A name to identify the network-server."
              required={true}
            />
            <TextField
              id="server"
              label="Network-server server"
              fullWidth={true}
              margin="normal"
              value={this.state.object.server || ""}
              onChange={this.onChange}
              helperText="The 'hostname:port' of the network-server, e.g. 'localhost:8000'."
              required={true}
            />
          </div>}
          {this.state.tab === 1 && <div>
            <FormControl
              label="Gateway discovery"
            >
              <FormGroup>
                <FormControlLabel
                  control={
                    <Checkbox
                      id="gatewayDiscoveryEnabled"
                      checked={!!this.state.object.gatewayDiscoveryEnabled}
                      onChange={this.onChange}
                      value="true"
                      color="primary"
                    />
                  }
                  label="Enable gateway discovery"
                />
              </FormGroup>
              <FormHelperText>Enable the gateway discovery feature for this network-server.</FormHelperText>
            </FormControl>
            {this.state.object.gatewayDiscoveryEnabled && <TextField
              id="gatewayDiscoveryInterval"
              label="Interval (per day)"
              type="number"
              fullWidth={true}
              margin="normal"
              value={this.state.object.gatewayDiscoveryInterval}
              onChange={this.onChange}
              helperText="The number of gateway discovery 'pings' per day that ChirpStack Application Server will broadcast through each gateway."
              required={true}
            />}
            {this.state.object.gatewayDiscoveryEnabled && <TextField
              id="gatewayDiscoveryTXFrequency"
              label="TX frequency (Hz)"
              type="number"
              fullWidth={true}
              margin="normal"
              value={this.state.object.gatewayDiscoveryTXFrequency}
              onChange={this.onChange}
              helperText="The frequency (Hz) used for transmitting the gateway discovery 'pings'. Please consult the LoRaWAN Regional Parameters specification for the channels valid for each region."
              required={true}
            />}
            {this.state.object.gatewayDiscoveryEnabled && <TextField
              id="gatewayDiscoveryDR"
              label="TX data-rate"
              type="number"
              fullWidth={true}
              margin="normal"
              value={this.state.object.gatewayDiscoveryDR}
              onChange={this.onChange}
              helperText="The data-rate used for transmitting the gateway discovery 'pings'. Please consult the LoRaWAN Regional Parameters specification for the data-rates valid for each region."
              required={true}
            />}
          </div>}
          {this.state.tab === 2 && <div>
            <FormControl
              label="Certificates for ChirpStack Application Server to ChirpStack Network Server connection"
            >
              <FormGroup>
                <TextField
                  id="caCert"
                  label="CA certificate"
                  fullWidth={true}
                  margin="normal"
                  value={this.state.object.caCert || ""}
                  onChange={this.onChange}
                  helperText="Paste the content of the CA certificate (PEM) file in the above textbox. Leave blank to disable TLS."
                  multiline
                  rows="4"
                />
                <TextField
                  id="tlsCert"
                  label="TLS certificate"
                  fullWidth={true}
                  margin="normal"
                  value={this.state.object.tlsCert || ""}
                  onChange={this.onChange}
                  helperText="Paste the content of the TLS certificate (PEM) file in the above textbox. Leave blank to disable TLS."
                  multiline
                  rows="4"
                />
                <TextField
                  id="tlsKey"
                  label="TLS key"
                  fullWidth={true}
                  margin="normal"
                  value={this.state.object.tlsKey || ""}
                  onChange={this.onChange}
                  helperText="Paste the content of the TLS key (PEM) file in the above textbox. Leave blank to disable TLS. Note: for security reasons, the TLS key can't be retrieved after being submitted (the field is left blank). When re-submitting the form with an empty TLS key field (but populated TLS certificate field), the key won't be overwritten."
                  multiline
                  rows="4"
                />
              </FormGroup>
            </FormControl>

            <FormControl
              label="Certificates for ChirpStack Network Server to ChirpStack Application Server connection"
            >
              <FormGroup>
                <TextField
                  id="routingProfileCACert"
                  label="CA certificate"
                  fullWidth={true}
                  margin="normal"
                  value={this.state.object.routingProfileCACert || ""}
                  onChange={this.onChange}
                  helperText="Paste the content of the CA certificate (PEM) file in the above textbox. Leave blank to disable TLS."
                  multiline
                  rows="4"
                />
                <TextField
                  id="routingProfileTLSCert"
                  label="TLS certificate"
                  fullWidth={true}
                  margin="normal"
                  value={this.state.object.routingProfileTLSCert || ""}
                  onChange={this.onChange}
                  helperText="Paste the content of the TLS certificate (PEM) file in the above textbox. Leave blank to disable TLS."
                  multiline
                  rows="4"
                />
                <TextField
                  id="routingProfileTLSKey"
                  label="TLS key"
                  fullWidth={true}
                  margin="normal"
                  value={this.state.object.routingProfileTLSKey || ""}
                  onChange={this.onChange}
                  helperText="Paste the content of the TLS key (PEM) file in the above textbox. Leave blank to disable TLS. Note: for security reasons, the TLS key can't be retrieved after being submitted (the field is left blank). When re-submitting the form with an empty TLS key field (but populated TLS certificate field), the key won't be overwritten."
                  multiline
                  rows="4"
                />
              </FormGroup>
            </FormControl>
          </div>}
      </Form>
    );
  }
}

export default NetworkServerForm;
