import React, { Component } from 'react';
import { Link } from 'react-router';

import { Map, Marker, TileLayer } from 'react-leaflet';
import Select from "react-select";

import SessionStore from "../stores/SessionStore";
import OrganizationStore from "../stores/OrganizationStore";
import LocationStore from "../stores/LocationStore";
import GatewayStore from "../stores/GatewayStore";


class GatewayForm extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      isGlobalAdmin: false,
      gateway: {},
      mapZoom: 15,
      initialOrganizationOptions: [],
      macDisabled: false,
      channelConfigurations: [],
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.updatePosition = this.updatePosition.bind(this);
    this.updateZoom = this.updateZoom.bind(this);
    this.setToCurrentPosition = this.setToCurrentPosition.bind(this);
    this.handleSetToCurrentPosition = this.handleSetToCurrentPosition.bind(this);
    this.onOrganizationAutocomplete = this.onOrganizationAutocomplete.bind(this);
    this.onOrganizationSelect = this.onOrganizationSelect.bind(this);
    this.onChannelConfigurationChange = this.onChannelConfigurationChange.bind(this);
    this.setSelectedOrganization = this.setSelectedOrganization.bind(this);
    this.setInitialOrganizations = this.setInitialOrganizations.bind(this);
  }

  onChange(field, e) {
    let gateway = this.state.gateway;

    if (e.target.type === "number") {
      gateway[field] = parseFloat(e.target.value);
    } else if (e.target.type === "checkbox") {
      gateway[field] = e.target.checked;
    } else {
      gateway[field] = e.target.value;
    }

    this.setState({
      gateway: gateway,
    });
  }

  updatePosition() {
    const position = this.refs.marker.leafletElement.getLatLng();
    let gateway = this.state.gateway;
    gateway.latitude = position.lat;
    gateway.longitude = position.lng;
    this.setState({
      gateway: gateway,
    });
  }

  updateZoom(e) {
    this.setState({
      mapZoom: e.target.getZoom(),
    });
  }

  componentDidMount() {
    this.setState({
      gateway: this.props.gateway,
      isGlobalAdmin: SessionStore.isAdmin(),
    }, () => {
      this.setSelectedOrganization();
    });

    if (!this.props.update) { 
      this.setToCurrentPosition(false);
    }

    GatewayStore.getAllChannelConfigurations((configurations) => {
      this.setState({
        channelConfigurations: configurations,
      });
    });

    SessionStore.on("change", () => {
      this.setState({
        isGlobalAdmin: SessionStore.isAdmin(),
      });
    });
  }

  setToCurrentPosition(overwrite) {
    LocationStore.getLocation((position) => {
      if (overwrite === true || typeof(this.state.gateway.latitude) === "undefined" || typeof(this.state.gateway.longitude) === "undefined" || this.state.gateway.latitude === 0 || this.state.gateway.longitude === 0) {
        let gateway = this.state.gateway;
        gateway.latitude = position.coords.latitude;
        gateway.longitude = position.coords.longitude;
        this.setState({
          gateway: gateway,
        });
      }
    });
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      gateway: nextProps.gateway, 
      macDisabled: typeof nextProps.gateway.mac !== "undefined",
    }, () => {
      this.setSelectedOrganization();
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.gateway);
  }

  handleSetToCurrentPosition(e) {
    e.preventDefault();
    this.setToCurrentPosition(true);
  }

  onOrganizationAutocomplete(input, callbackFunc) {
    OrganizationStore.getAll(input, 10, 0, (totalCount, orgs) => {
      const options = orgs.map((org, i) => {
        return {
          value: org.id,
          label: org.displayName,
        };
      });

      callbackFunc(null, {
        options: options,
        complete: true,
      });
    });
  }

  onOrganizationSelect(val) {
    let gateway = this.state.gateway;
    gateway.organizationID = val.value;
    this.setState({
      gateway: gateway,
      initialOrganizationOptions: [val],
    });
  }

  onChannelConfigurationChange(val) {
    let gateway = this.state.gateway;
    if (val != null) {
      gateway.channelConfigurationID = val.value;
    } else {
      gateway.channelConfigurationID = null;
    }
    this.setState({
      gateway: gateway,
    });
  }

  setSelectedOrganization() {
    if (typeof(this.state.gateway.organizationID) === "undefined") {
      return;
    }
    OrganizationStore.getOrganization(this.state.gateway.organizationID, (org) => {
      this.setState({
        initialOrganizationOptions: [{
          value: org.id,
          label: org.displayName,
        }],
      });
    });
  }

  setInitialOrganizations() {
    OrganizationStore.getAll("", 10, 0, (totalCount, orgs) => {
      const options = orgs.map((org, i) => {
        return {
          value: org.id,
          label: org.displayName,
        };
      });

      this.setState({
        initialOrganizationOptions: options,
      });
    });
  }

  render() {
    const mapStyle = {
      height: "400px",
    };

    let position = [];

    if (typeof(this.state.gateway.latitude) !== "undefined" || typeof(this.state.gateway.longitude) !== "undefined") {
      position = [this.state.gateway.latitude, this.state.gateway.longitude];
    } else {
      position = [0,0];
    }

    const channelConfigurations = this.state.channelConfigurations.map((c, i) => {
      return {
        value: c.id,
        label: c.name,
      };
    });

    return(
      <div>
        <form onSubmit={this.handleSubmit}>
          <div className="form-group">
            <label className="control-label" htmlFor="name">Gateway name</label>
            <input className="form-control" id="name" type="text" placeholder="e.g. 'rooftop-gateway'" required value={this.state.gateway.name || ''} pattern="[\w-]+" onChange={this.onChange.bind(this, 'name')} />
            <p className="help-block">
              The name may only contain words, numbers and dashes.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="name">Gateway description</label>
            <input className="form-control" id="description" type="text" placeholder="a short description of your gateway" required value={this.state.gateway.description || ''} onChange={this.onChange.bind(this, 'description')} />
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="mac">MAC address</label>
            <input className="form-control" id="mac" type="text" placeholder="0000000000000000" pattern="[A-Fa-f0-9]{16}" required disabled={this.state.macDisabled} value={this.state.gateway.mac || ''} onChange={this.onChange.bind(this, 'mac')} /> 
            <p className="help-block">
              Enter the gateway MAC address as configured in the packet-forwarder configuration on the gateway.
            </p>
          </div>
          <div className={"form-group " + (this.state.isGlobalAdmin && this.props.update ? '' : 'hidden')}>
            <label className="control-label" htmlFor="organization">Organization</label>
            <Select.Async
              name="organization"
              required
              options={this.state.initialOrganizationOptions}
              loadOptions={this.onOrganizationAutocomplete}
              value={this.state.gateway.organizationID}
              onChange={this.onOrganizationSelect}
              clearable={false}
              autoload={false}
              onOpen={this.setInitialOrganizations}
            /> 
            <p className="help-block">Note that moving a gateway to a different organization can only be done by global admin users.</p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="channelConfigurationID">Channel-configuration</label>
            <Select
              name="channelConfigurationID"
              options={channelConfigurations}
              value={this.state.gateway.channelConfigurationID}
              onChange={this.onChannelConfigurationChange}
            />
            <p className="help-block">An optional channel-configuration can be assigned to a gateway. This configuration can be used to automatically re-configure the gateway (in the future).</p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="ping">
              <input type="checkbox" name="ping" id="ping" checked={this.state.gateway.ping} onChange={this.onChange.bind(this, 'ping')} /> Ping enabled
            </label>
            <p className="help-block">When enabled (and LoRa App Server is configured to send gateway pings), the gateway will send out periodical pings to test its coverage by other gateways in the same network.</p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="altitude">Gateway altitude (meters)</label>
            <input className="form-control" id="altitude" type="number" value={this.state.gateway.altitude || 0} onChange={this.onChange.bind(this, 'altitude')} />
            <p className="help-block">When the gateway has an on-board GPS, this value will be set automatically when the network received statistics from the gateway.</p>
          </div>
          <div className="form-group">
            <label className="control-label">Gateway location (<Link onClick={this.handleSetToCurrentPosition} href="#">set to current location</Link>)</label>
            <Map
              zoom={this.state.mapZoom}
              center={position}
              style={mapStyle}
              animate={true}
              onZoomend={this.updateZoom}
              scrollWheelZoom={false}
            >
              <TileLayer
                url='//{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
                attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
              />
              <Marker position={position} draggable={true} onDragend={this.updatePosition} ref="marker" />
            </Map>
            <p className="help-block">Drag the marker to the location of the gateway. When the gateway has an on-board GPS, this value will be set automatically when the network receives statistics from the gateway.</p>
          </div>
          <hr />
          <div className="btn-toolbar pull-right">
            <a className="btn btn-default" onClick={this.context.router.goBack}>Go back</a>
            <button type="submit" className="btn btn-primary">Submit</button>
          </div>
        </form>
      </div>
    );
  }
}

export default GatewayForm;
