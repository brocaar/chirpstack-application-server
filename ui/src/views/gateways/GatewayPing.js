import React, { Component } from 'react';

import moment from "moment";
import { Map, Marker, TileLayer, Polyline, Popup } from 'react-leaflet';

import GatewayStore from "../../stores/GatewayStore";


class GatewayPing extends Component {
  constructor() {
    super();

    this.state = {};
  }

  componentDidMount() {
    GatewayStore.getGateway(this.props.params.mac, (gateway) => {
      this.setState({
        gateway: gateway,
      })
    });

    GatewayStore.getLastPing(this.props.params.mac, (ping) => {
      this.setState({
        ping: ping,
      });
    });
  }

  render() {
    const style = {
      height: "800px",
    };

    if (!this.state.gateway || !this.state.ping || !this.state.ping.pingRX || this.state.ping.pingRX.length === 0) {
      return(
        <div className="panel panel-default">
          <div className="panel-body">
            No gateway ping data available (yet). This could mean:

            <ul>
              <li>no ping was emitted yet</li>
              <li>the gateway ping feature has been disabled in LoRa App Server</li>
              <li>the ping was not received by any other gateways</li>
            </ul>
          </div>
        </div>
      );
    }

    const lastPingTimestamp = moment(this.state.ping.createdAt).fromNow();

    let bounds = [];
    let markers = [];
    let lines = [];

    const gwPos = [this.state.gateway.latitude, this.state.gateway.longitude];
    markers.push(<Marker position={gwPos} key={"gw" + this.state.gateway.mac}>
      <Popup>
        <span>
          {this.state.gateway.mac}<br />
          Freq: {this.state.ping.frequency/1000000} MHz<br />
          DR: {this.state.ping.dr}<br />
          Altitude: {this.state.gateway.altitude} meter(s)
        </span>
      </Popup>
    </Marker>);

    bounds.push(gwPos);

    for (let rx of this.state.ping.pingRX) {
      const pingPos = [rx.latitude, rx.longitude];

      markers.push(<Marker position={pingPos} key={"ping" + rx.mac}>
        <Popup>
          <span>
            {rx.mac}<br />
            RSSI: {rx.rssi}<br />
            SNR: {rx.loraSNR}<br />
            Altitude: {rx.altitude} meter(s)
          </span>
        </Popup>
      </Marker>);
      bounds.push(pingPos);

      let color = "";
      if (rx.rssi >= -100) {
        color = "#FF0000";
      } else if (rx.rssi >= -105) {
        color = "#FF7F00";
      } else if (rx.rssi >= -110) {
        color = "#FFFF00";
      } else if (rx.rssi >= -115) {
        color = "#00FF00";
      } else if (rx.rssi >= -120) {
        color = "#00FFFF";
      } else {
        color = "#0000FF";
      }

      lines.push(<Polyline key={"line" + rx.mac} positions={[gwPos, pingPos]} color={color} opacity="0.7" weight="3" />);
    }


    return(
      <div className="panel panel-default">
        <div className="panel-heading">
        <h3 className="panel-title">Last ping: {lastPingTimestamp}</h3>
        </div>
        <div className="panel-body">
          <Map animate={true} style={style} maxZoom={19} scrollWheelZoom={false} bounds={bounds}>
            <TileLayer
              url='//{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
              attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
            />
            {markers}
            {lines}
          </Map>
        </div>
      </div>
    );
  }
}

export default GatewayPing;