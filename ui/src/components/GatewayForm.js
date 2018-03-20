import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';

import { Map, Marker, TileLayer } from 'react-leaflet';
import Select from "react-select";

import Loaded from "./Loaded.js";
import SessionStore from "../stores/SessionStore";
import LocationStore from "../stores/LocationStore";
import GatewayStore from "../stores/GatewayStore";
import NetworkServerStore from "../stores/NetworkServerStore";


class GatewayForm extends Component {
  constructor() {
    super();

    this.state = {
      gateway: {},
      mapZoom: 15,
      update: false,
      channelConfigurations: [],
      networkServers: [],
      loaded: {
        networkServers: false,
      },
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.updatePosition = this.updatePosition.bind(this);
    this.updateZoom = this.updateZoom.bind(this);
    this.setToCurrentPosition = this.setToCurrentPosition.bind(this);
    this.handleSetToCurrentPosition = this.handleSetToCurrentPosition.bind(this);
  }

  onSelectChange(field, val) {
    let gateway = this.state.gateway;
    if (val != null) {
      gateway[field] = val.value;
    } else {
      gateway[field] = null;
    }

    if (field === "networkServerID" && gateway.networkServerID !== null) {
      GatewayStore.getAllChannelConfigurations(gateway.networkServerID, (configurations) => {
        this.setState({
          channelConfigurations: configurations,
        });
      });
    }

    this.setState({
      gateway: gateway,
    });
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
    });

    if (!this.props.update) { 
      this.setToCurrentPosition(false);
    }

    NetworkServerStore.getAllForOrganizationID(this.props.organizationID, 9999, 0, (totalCount, networkServers) => {
      this.setState({
        networkServers: networkServers,
        loaded: {
          networkServers: true,
        },
      });
    });

    if (this.props.gateway.networkServerID !== undefined) {
      GatewayStore.getAllChannelConfigurations(this.props.gateway.networkServerID, (configurations) => {
        this.setState({
          channelConfigurations: configurations,
        });
      });
    }
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
      update: typeof nextProps.gateway.mac !== "undefined",
    });

    if (this.props.gateway.networkServerID !== undefined) {
      GatewayStore.getAllChannelConfigurations(nextProps.gateway.networkServerID, (configurations) => {
        this.setState({
          channelConfigurations: configurations,
        });
      });
    }
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.gateway);
  }

  handleSetToCurrentPosition(e) {
    e.preventDefault();
    this.setToCurrentPosition(true);
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

    const networkServerOptions = this.state.networkServers.map((n, i) => {
      return {
        value: n.id,
        label: n.name,
      };
    });

    return(
      <Loaded loaded={this.state.loaded}>
        <form onSubmit={this.handleSubmit}>
          <div className={"alert alert-warning " + (this.state.networkServers.length > 0 ? 'hidden' : '')}>
            No network-servers are associated with this organization, a <Link to={`/organizations/${this.props.organizationID}/service-profiles`}>service-profile</Link> needs to be created first for this organization.
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="name">Gateway name</label>
            <input className="form-control" id="name" type="text" placeholder="e.g. 'rooftop-gateway'" required value={this.state.gateway.name || ''} pattern="[\w-]+" onChange={this.onChange.bind(this, 'name')} />
            <p className="help-block">
              The name may only contain words, numbers and dashes.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="name">Gateway description</label>
            <textarea className="form-control" id="description" rows="4" placeholder="an optional note about the gateway" value={this.state.gateway.description || ''} onChange={this.onChange.bind(this, 'description')} />
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="mac">MAC address</label>
            <input className="form-control" id="mac" type="text" placeholder="0000000000000000" pattern="[A-Fa-f0-9]{16}" required disabled={this.state.update} value={this.state.gateway.mac || ''} onChange={this.onChange.bind(this, 'mac')} /> 
            <p className="help-block">
              Enter the gateway MAC address as configured in the packet-forwarder configuration on the gateway.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="networkServerID">Network-server</label>
            <Select
              name="networkServerID"
              options={networkServerOptions}
              value={this.state.gateway.networkServerID}
              onChange={this.onSelectChange.bind(this, "networkServerID")}
              disabled={this.state.update}
            />
            <p className="help-block">
              Select the network-server to which the gateway will connect. When no network-servers are available in the dropdown, make sure a service-profile exists for this organization. 
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="channelConfigurationID">Channel-configuration</label>
            <Select
              name="channelConfigurationID"
              options={channelConfigurations}
              value={this.state.gateway.channelConfigurationID}
              onChange={this.onSelectChange.bind(this, "channelConfigurationID")}
            />
            <p className="help-block">An optional channel-configuration can be assigned to a gateway. This configuration can be used to automatically re-configure the gateway (in the future).</p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="ping">
              <input type="checkbox" name="ping" id="ping" checked={!!this.state.gateway.ping} onChange={this.onChange.bind(this, 'ping')} /> Discovery enabled
            </label>
            <p className="help-block">When enabled (and LoRa App Server is configured with the gateway discover feature enabled), the gateway will send out periodical pings to test its coverage by other gateways in the same network.</p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="altitude">Gateway altitude (meters)</label>
            <input className="form-control" id="altitude" type="number" value={this.state.gateway.altitude || 0} onChange={this.onChange.bind(this, 'altitude')} />
            <p className="help-block">When the gateway has an on-board GPS, this value will be set automatically when the network received statistics from the gateway.</p>
          </div>
          <div className="form-group">
            <label className="control-label">Gateway location (<a onClick={this.handleSetToCurrentPosition} href="#getLocation">set to current location</a>)</label>
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
            <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
            <button type="submit" className="btn btn-primary">Submit</button>
          </div>
        </form>
      </Loaded>
    );
  }
}

export default withRouter(GatewayForm);
