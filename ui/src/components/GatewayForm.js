import React, { Component } from 'react';
import { Link } from 'react-router';

import { Map, Marker, TileLayer } from 'react-leaflet';

class GatewayForm extends Component {
  constructor() {
    super();

    this.state = {
      gateway: {},
      mapZoom: 15,
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.updatePosition = this.updatePosition.bind(this);
    this.updateZoom = this.updateZoom.bind(this);
    this.setToCurrentPosition = this.setToCurrentPosition.bind(this);
    this.handleSetToCurrentPosition = this.handleSetToCurrentPosition.bind(this);
  }

  onChange(field, e) {
    let gateway = this.state.gateway;

    if (e.target.type === "number") {
      gateway[field] = parseFloat(e.target.value);
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

  componentWillMount() {
    this.setState({
      gateway: this.props.gateway,
    });

    this.setToCurrentPosition(false);
  }

  setToCurrentPosition(overwrite) {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
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
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      gateway: nextProps.gateway, 
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
            <input className="form-control" id="mac" type="text" placeholder="0000000000000000" pattern="[A-Fa-f0-9]{16}" required value={this.state.gateway.mac || ''} onChange={this.onChange.bind(this, 'mac')} /> 
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="altitude">Gateway altitude</label>
            <input className="form-control" id="altitude" type="number" step="0.01" value={this.state.gateway.altitude || 0} onChange={this.onChange.bind(this, 'altitude')} />
            <p className="help-block">When the gateway has an on-board GPS, this value will be set automatically when the network received statistics from the gateway.</p>
          </div>
          <div className="form-group">
            <label className="control-label">Gateway location (<Link onClick={this.handleSetToCurrentPosition} href="#">set to current location</Link>)</label>
            <Map zoom={this.state.mapZoom} center={position} style={mapStyle} animate={true} onZoomend={this.updateZoom}>
              <TileLayer
                url='http://{s}.tile.osm.org/{z}/{x}/{y}.png'
                attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
              />
              <Marker position={position} draggable={true} onDragend={this.updatePosition} ref="marker" />
            </Map>
            <p className="help-block">Drag the marker to the location of the gateway. When the gateway has an on-board GPS, this value will be set automatically when the network receives statistics from the gateway.</p>
          </div>
          <hr />
          <button type="submit" className="btn btn-primary pull-right">Submit</button>
        </form>
      </div>
    );
  }
}

export default GatewayForm;
