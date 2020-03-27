import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Paper from '@material-ui/core/Paper';
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import Grid from '@material-ui/core/Grid';

import moment from "moment";
import { Map, Marker } from 'react-leaflet';
import { Line } from "react-chartjs-2";

import MapTileLayer from "../../components/MapTileLayer";
import GatewayStore from "../../stores/GatewayStore";


const styles = {
  chart: {
    height: 300,
  },
};


class GatewayDetails extends Component {
  constructor() {
    super();
    this.state = {};
    this.loadStats = this.loadStats.bind(this);
  }

  componentDidMount() {
    this.loadStats();
  }

  loadStats() {
    const end = moment().toISOString()
    const start = moment().subtract(30, "days").toISOString()

    GatewayStore.getStats(this.props.match.params.gatewayID, start, end, resp => {
      let statsDown = {
        labels: [],
        datasets: [
          {
            label: "rx received",
            borderColor: "rgba(33, 150, 243, 1)",
            backgroundColor: "rgba(0, 0, 0, 0)",
            lineTension: 0,
            pointBackgroundColor: "rgba(33, 150, 243, 1)",
            data: [],
          },
        ],
      }

      let statsUp = {
        labels: [],
        datasets: [
          {
            label: "tx emitted",
            borderColor: "rgba(33, 150, 243, 1)",
            backgroundColor: "rgba(0, 0, 0, 0)",
            lineTension: 0,
            pointBackgroundColor: "rgba(33, 150, 243, 1)",
            data: [],
          },
        ],
      }

      for (const row of resp.result) {
        statsUp.labels.push(moment(row.timestamp).format("Do"));
        statsDown.labels.push(moment(row.timestamp).format("Do"));
        statsUp.datasets[0].data.push(row.txPacketsEmitted);
        statsDown.datasets[0].data.push(row.rxPacketsReceivedOK);
      }

      this.setState({
        statsUp: statsUp,
        statsDown: statsDown,
      });
    });
  }

  render() {
    if (this.props.gateway === undefined || this.state.statsDown === undefined || this.state.statsUp === undefined) {
      return(<div></div>);
    }

    const style = {
      height: 400,
    };

    const statsOptions = {
      legend: {
        display: false,
      },
      maintainAspectRatio: false,
      scales: {
        yAxes: [{
          ticks: {
            beginAtZero: true,
          },
        }],
      },
    }

    let position = [];
    if (typeof(this.props.gateway.location.latitude) !== "undefined" && typeof(this.props.gateway.location.longitude !== "undefined")) {
      position = [this.props.gateway.location.latitude, this.props.gateway.location.longitude]; 
    } else {
      position = [0,0];
    }

    let lastseen = "Never";
    if (this.props.lastSeenAt !== null) {
      lastseen = moment(this.props.lastSeenAt).fromNow();
    }

    return(
      <Grid container spacing={4}>
        <Grid item xs={6}>
          <Card>
            <CardHeader
              title="Gateway details"
            />
            <CardContent>
              <Typography variant="subtitle1" color="primary">
                Gateway ID
              </Typography>
              <Typography variant="body1" gutterBottom>
                {this.props.gateway.id}
              </Typography>
              <Typography variant="subtitle1" color="primary">
                Altitude
              </Typography>
              <Typography variant="body1" gutterBottom>
                {this.props.gateway.location.altitude} meters
              </Typography>
              <Typography variant="subtitle1" color="primary">
                GPS coordinates
              </Typography>
              <Typography variant="body1" gutterBottom>
                {this.props.gateway.location.latitude}, {this.props.gateway.location.longitude}
              </Typography>
              <Typography variant="subtitle1" color="primary">
                Last seen
              </Typography>
              <Typography variant="body1" gutterBottom>
                {lastseen}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Paper>
            <Map center={position} zoom={15} style={style} animate={true} scrollWheelZoom={false}>
              <MapTileLayer />
              <Marker position={position} />
            </Map>
          </Paper>
        </Grid>
        <Grid item xs={12}>
          <Card>
            <CardHeader title="Frames received" />
            <CardContent className={this.props.classes.chart}>
              <Line height={75} options={statsOptions} data={this.state.statsDown} redraw />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12}>
          <Card>
            <CardHeader title="Frames transmitted" />
            <CardContent className={this.props.classes.chart}>
              <Line height={75} options={statsOptions} data={this.state.statsUp} redraw />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(GatewayDetails);
