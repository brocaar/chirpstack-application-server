import React, { Component } from 'react';
import moment from "moment";
import { Map, Marker, TileLayer } from 'react-leaflet';
import { Bar } from "react-chartjs";

import GatewayStore from "../../stores/GatewayStore";


class GatewayStats extends Component {
  constructor() {
    super();

    this.state = {
      periodSelected: '30d',
      periods: {
        "hour": {
          interval: "MINUTE",
          substract: 59,
          substractInterval: 'minutes',
          format: "mm",
        },
        "1d": {
          interval: "HOUR",
          substract: 23,
          substractInterval: "hours",
          format: "HH",
        },
        "14d": {
          interval: "DAY",
          substract: 13,
          substractInterval: "days",
          format: "Do",
        },
        "30d": {
          interval: "DAY",
          substract: 29,
          substractInterval: "days",
          format: "Do",
        },
      },
      statsUp: {
        labels: [],
        datasets: [
          {
            label: "received for transmission",
            data: [],
            fillColor: "rgba(33, 150, 243, 0.25)",
          },
          {
            label: "emitted",
            data: [],
            fillColor: "rgba(33, 150, 243, 1)",
          },
        ],
      },
      statsDown: {
        labels: [],
        datasets: [
          {
            label: "total received",
            data: [],
            fillColor: "rgba(33, 150, 243, 0.25)",
          },
          {
            label: "received with valid CRC",
            data: [],
            fillColor: "rgba(33, 150, 243, 1)",
          },
        ],
      },
      statsOptions: {
        animation: true,
        barShowStroke: false,
        barValueSpacing: 4,
        responsive: true,
      },
    };

    this.updateStats = this.updateStats.bind(this);
  }

  componentWillMount() {
    this.updateStats(this.state.periodSelected);
  }

  updateStats(period) {
    GatewayStore.getGatewayStats(this.props.mac, this.state.periods[period].interval, moment().subtract(this.state.periods[period].substract, this.state.periods[period].substractInterval).toISOString(), moment().toISOString(), (records) => {
      let statsUp = this.state.statsUp;
      let statsDown = this.state.statsDown;

      statsUp.labels = [];
      statsDown.labels = [];
      statsUp.datasets[0].data = [];
      statsUp.datasets[1].data = [];
      statsDown.datasets[0].data = [];
      statsDown.datasets[1].data = [];

      for (const record of records) {
        statsUp.labels.push(moment(record.timestamp).format(this.state.periods[period].format));
        statsDown.labels.push(moment(record.timestamp).format(this.state.periods[period].format));
        statsUp.datasets[0].data.push(record.txPacketsReceived);
        statsUp.datasets[1].data.push(record.txPacketsEmitted);
        statsDown.datasets[0].data.push(record.rxPacketsReceived);
        statsDown.datasets[1].data.push(record.rxPacketsReceivedOK);
      }

      this.setState({
        statsUp: statsUp,
        statsDown: statsDown,
      });
    });  
  }

  updatePeriod(p) {
    this.setState({
      periodSelected: p,
    });

    this.updateStats(p);
  }


  render() {
    return(
      <div>
        <div className="clearfix">
          <div className="btn-group pull-right" role="group" aria-label="...">
            <button type="button" className={'btn btn-' + (this.state.periodSelected === 'hour' ? 'primary' : 'default')} onClick={this.updatePeriod.bind(this, 'hour')}>hour</button>
            <button type="button" className={'btn btn-' + (this.state.periodSelected === '1d' ? 'primary' : 'default')} onClick={this.updatePeriod.bind(this, '1d')}>1D</button>
            <button type="button" className={'btn btn-' + (this.state.periodSelected === '14d' ? 'primary' : 'default')} onClick={this.updatePeriod.bind(this, '14d')}>14D</button>
            <button type="button" className={'btn btn-' + (this.state.periodSelected === '30d' ? 'primary' : 'default')} onClick={this.updatePeriod.bind(this, '30d')}>30D</button>
          </div>
        </div>

        <h4>Frames sent per {this.state.periods[this.state.periodSelected].interval.toLowerCase()}</h4>
        <Bar height="75" data={this.state.statsUp} options={this.state.statsOptions} redraw />
        <hr />
        <h4>Frames received per {this.state.periods[this.state.periodSelected].interval.toLowerCase()}</h4>
        <Bar height="75" data={this.state.statsDown} options={this.state.statsOptions} redraw /> 


      </div>
    );
  }
}

class GatewayDetails extends Component {
  constructor() {
    super();

    this.state = {
      gateway: {},
    }
  }

  componentWillMount() {
    GatewayStore.getGateway(this.props.match.params.mac, (gateway) => {
      this.setState({
        gateway: gateway,
      });
    }); 
  }

  render() {
    const style = {
      height: "400px",
    };

    let lastseen = "";
    let position = [];

    if (typeof(this.state.gateway.latitude) !== "undefined" && typeof(this.state.gateway.longitude !== "undefined")) {
      position = [this.state.gateway.latitude, this.state.gateway.longitude]; 
    } else {
      position = [0,0];
    }

    if (typeof(this.state.gateway.lastSeenAt) !== "undefined" && this.state.gateway.lastSeenAt !== "") {
      lastseen = moment(this.state.gateway.lastSeenAt).fromNow();    
    }

    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <div className="row">
            <div className="col-md-6">
              <table className="table">
                <thead>
                  <tr>
                    <th colSpan={2}><h4>{this.state.gateway.name}</h4></th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td className="col-md-4"><strong>MAC</strong></td>
                    <td>{this.state.gateway.mac}</td>
                  </tr>
                  <tr>
                    <td className="col-md-4"><strong>Altitude</strong></td>
                    <td>{this.state.gateway.altitude} meters</td>
                  </tr>
                  <tr>
                    <td className="col-md-4"><strong>GPS coordinates</strong></td>
                    <td>{this.state.gateway.latitude}, {this.state.gateway.longitude}</td>
                  </tr>
                  <tr>
                    <td className="col-md-4"><strong>Last seen (stats)</strong></td>
                    <td>{lastseen}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div className="col-md-6">
              <Map center={position} zoom={15} style={style} animate={true} scrollWheelZoom={false}>
                <TileLayer
                  url='//{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
                  attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
                />
                <Marker position={position} />
              </Map>
            </div>
          </div>
          <hr />
          <GatewayStats mac={this.props.match.params.mac} />
        </div>
      </div>
    );
  }
}

export default GatewayDetails;
