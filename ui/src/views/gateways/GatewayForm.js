import React from "react";

import { withStyles } from "@material-ui/core/styles";
import TextField from '@material-ui/core/TextField';
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormHelperText from "@material-ui/core/FormHelperText";
import FormGroup from "@material-ui/core/FormGroup";
import FormLabel from "@material-ui/core/FormLabel";
import Checkbox from '@material-ui/core/Checkbox';

import { Map, Marker } from 'react-leaflet';

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import AutocompleteSelect from "../../components/AutocompleteSelect";
import NetworkServerStore from "../../stores/NetworkServerStore";
import GatewayProfileStore from "../../stores/GatewayProfileStore";
import LocationStore from "../../stores/LocationStore";
import MapTileLayer from "../../components/MapTileLayer";
import theme from "../../theme";


const styles = {
  mapLabel: {
    marginBottom: theme.spacing.unit,
  },
  link: {
    color: theme.palette.primary.main,
  },
  formLabel: {
    fontSize: 12,
  },
};


class GatewayForm extends FormComponent {
  constructor() {
    super();
    
    this.state = {
      mapZoom: 15,
    };

    this.getNetworkServerOption = this.getNetworkServerOption.bind(this);
    this.getNetworkServerOptions = this.getNetworkServerOptions.bind(this);
    this.getGatewayProfileOption = this.getGatewayProfileOption.bind(this);
    this.getGatewayProfileOptions = this.getGatewayProfileOptions.bind(this);
    this.setCurrentPosition = this.setCurrentPosition.bind(this);
    this.updatePosition = this.updatePosition.bind(this);
    this.updateZoom = this.updateZoom.bind(this);
  }

  componentDidMount() {
    super.componentDidMount();

    if (!this.props.update) {
      this.setCurrentPosition();
    }
  }

  onChange(e) {
    if (e.target.id === "networkServerID" && e.target.value !== this.state.object.networkServerID) {
      let object = this.state.object;
      object.gatewayProfileID = null;
      this.setState({
        object: object,
      });
    }

    super.onChange(e);
  }

  setCurrentPosition(e) {
    if (e !== undefined) {
      e.preventDefault();
    }

    LocationStore.getLocation(position => {
      let object = this.state.object;
      object.location = {
        latitude: position.coords.latitude,
        longitude: position.coords.longitude,
      }
      this.setState({
        object: object,
      });
    });
  }

  updatePosition() {
    const position = this.refs.marker.leafletElement.getLatLng();
    let object = this.state.object;
    object.location = {
      latitude: position.lat,
      longitude: position.lng,
    }
    this.setState({
      object: object,
    });
  }

  updateZoom(e) {
    this.setState({
      mapZoom: e.target.getZoom(),
    });
  }

  getNetworkServerOption(id, callbackFunc) {
    NetworkServerStore.get(id, resp => {
      callbackFunc({label: resp.networkServer.name, value: resp.networkServer.id});
    });
  }

  getNetworkServerOptions(search, callbackFunc) {
    NetworkServerStore.list(this.props.match.params.organizationID, 999, 0, resp => {
      const options = resp.result.map((ns, i) => {return {label: ns.name, value: ns.id}});
      callbackFunc(options);
    });
  }

  getGatewayProfileOption(id, callbackFunc) {
    GatewayProfileStore.get(id, resp => {
      callbackFunc({label: resp.gatewayProfile.name, value: resp.gatewayProfile.id});
    });
  }

  getGatewayProfileOptions(search, callbackFunc) {
    if (this.state.object === undefined || this.state.object.networkServerID === undefined) {
      callbackFunc([]);
      return;
    }

    GatewayProfileStore.list(this.state.object.networkServerID, 999, 0, resp => {
      const options = resp.result.map((gp, i) => {return {label: gp.name, value: gp.id}});
      callbackFunc(options);
    });
  }

  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    const style = {
      height: 400,
    };

    let position = [];
    if (this.state.object.location.latitude !== undefined && this.state.object.location.longitude !== undefined) {
      position = [this.state.object.location.latitude, this.state.object.location.longitude];
    } else {
      position = [0, 0];
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <TextField
          id="name"
          label="Gateway name"
          margin="normal"
          value={this.state.object.name || ""}
          onChange={this.onChange}
          inputProps={{
            pattern: "[\\w-]+",
          }}
          helperText="The name may only contain words, numbers and dashes."
          required
          fullWidth
        />
        <TextField
          id="description"
          label="Gateway description"
          margin="normal"
          value={this.state.object.description || ""}
          onChange={this.onChange}
          rows={4}
          multiline
          required
          fullWidth
        />
        {!this.props.update && <TextField
          id="id"
          label="Gateway ID"
          margin="normal"
          value={this.state.object.id || ""}
          onChange={this.onChange}
          inputProps={{
            pattern: "[A-Fa-f0-9]{16}",
          }}
          placeholder="0000000000000000"
          required
          fullWidth
        />}
        {!this.props.update && <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel} required>Network-server</FormLabel>
          <AutocompleteSelect
            id="networkServerID"
            label="Select network-server"
            value={this.state.object.networkServerID || ""}
            onChange={this.onChange}
            getOption={this.getNetworkServerOption}
            getOptions={this.getNetworkServerOptions}
          />
          <FormHelperText>
            Select the network-server to which the gateway will connect. When no network-servers are available in the dropdown, make sure a service-profile exists for this organization. 
          </FormHelperText>
        </FormControl>}
        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.formLabel}>Gateway-profile</FormLabel>
          <AutocompleteSelect
            id="gatewayProfileID"
            label="Select gateway-profile"
            value={this.state.object.gatewayProfileID || ""}
            triggerReload={this.state.object.networkServerID || ""}
            onChange={this.onChange}
            getOption={this.getGatewayProfileOption}
            getOptions={this.getGatewayProfileOptions}
            inputProps={{
              clearable: true,
              cache: false,
            }}
          />
          <FormHelperText>
            An optional gateway-profile which can be assigned to a gateway. This configuration can be used to automatically re-configure the gateway when LoRa Gateway Bridge is configured so that it manages the packet-forwarder configuration.
          </FormHelperText>
        </FormControl>
        <FormGroup>
          <FormControlLabel
            label="Gateway discovery enabled"
            control={
              <Checkbox
                id="discoveryEnabled"
                checked={!!this.state.object.discoveryEnabled}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
          <FormHelperText>
            When enabled (and LoRa Server is configured with the gateway discover feature enabled), the gateway will send out periodical pings to test its coverage by other gateways in the same network.
          </FormHelperText>
        </FormGroup>
        <TextField
          id="location.altitude"
          label="Gateway altitude (meters)"
          margin="normal"
          type="number"
          value={this.state.object.location.altitude || 0}
          onChange={this.onChange}
          helperText="When the gateway has an on-board GPS, this value will be set automatically when the network received statistics from the gateway."
          required
          fullWidth
        />
        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.mapLabel}>Gateway location (<a onClick={this.setCurrentPosition} href="#getlocation" className={this.props.classes.link}>set to current location</a>)</FormLabel>
          <Map
            center={position}
            zoom={this.state.mapZoom}
            style={style}
            animate={true}
            scrollWheelZoom={false}
            onZoomend={this.updateZoom}
            >
            <MapTileLayer />
            <Marker position={position} draggable={true} onDragend={this.updatePosition} ref="marker" />
          </Map>
          <FormHelperText>
            Drag the marker to the location of the gateway. When the gateway has an on-board GPS, this value will be set automatically when the network receives statistics from the gateway.
          </FormHelperText>
        </FormControl>
      </Form>
    );
  }
}

export default withStyles(styles)(GatewayForm);
