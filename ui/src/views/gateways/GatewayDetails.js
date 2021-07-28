import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Paper from '@material-ui/core/Paper';
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import Grid from '@material-ui/core/Grid';
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";

import moment from "moment";
import { Map, Marker } from 'react-leaflet';
import { Line, Bar } from "react-chartjs-2";

import MapTileLayer from "../../components/MapTileLayer";
import GatewayStore from "../../stores/GatewayStore";
import Heatmap from "../../components/Heatmap";



const styles = {
  chart: {
    height: 300,
  },
};

class DetailsCard extends Component {
  render() {
    return(
      <Card>
        <CardHeader title="Gateway details" />
        <CardContent>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell>Gateway ID</TableCell>
                <TableCell>{this.props.gateway.id}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Altitude</TableCell>
                <TableCell>{this.props.gateway.location.altitude} meters</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>GPS coordinates</TableCell>
                <TableCell>{this.props.gateway.location.latitude}, {this.props.gateway.location.longitude}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Last seen at</TableCell>
                <TableCell>{this.props.lastSeenAt}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  }
}

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
      };

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
      };

      let statsDownError = {
        labels: [],
        datasets: [],
      };
      let statsDownErrorSets = {};

      let statsUpFreq = [];
      let statsDownFreq = [];
      let statsUpDr = [];
      let statsDownDr = [];

      for (const row of resp.result) {
        statsUp.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));
        statsUp.datasets[0].data.push(row.txPacketsEmitted);

        statsDown.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));
        statsDown.datasets[0].data.push(row.rxPacketsReceivedOK);

        statsDownError.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));
        Object.entries(row.txPacketsPerStatus).forEach(([k, v]) => {
          if (statsDownErrorSets[k] === undefined) {
            statsDownErrorSets[k] = [];
          }

          // fill gaps with 0s
          for (let i = statsDownErrorSets[k].length; i < statsDownError.labels.length - 1; i++) {
            statsDownErrorSets[k].push(0);
          }

          statsDownErrorSets[k].push(v);
        });

        statsUpFreq.push({
          x: moment(row.timestamp).format("YYYY-MM-DD"),
          y: row.rxPacketsPerFrequency,
        });

        statsDownFreq.push({
          x: moment(row.timestamp).format("YYYY-MM-DD"),
          y: row.txPacketsPerFrequency,
        });

        statsUpDr.push({
          x: moment(row.timestamp).format("YYYY-MM-DD"),
          y: row.rxPacketsPerDr,
        });

        statsDownDr.push({
          x: moment(row.timestamp).format("YYYY-MM-DD"),
          y: row.txPacketsPerDr,
        });
      }

      let backgroundColors = ['#8bc34a', '#ff5722', '#ff9800', '#ffc107', '#ffeb3b', '#cddc39', '#4caf50', '#009688', '#00bcd4', '#03a9f4', '#2196f3', '#3f51b5', '#673ab7', '#9c27b0', '#e91e63'];

      Object.entries(statsDownErrorSets).forEach(([k, v]) => {
        statsDownError.datasets.push({
          label: k,
          data: v,
          backgroundColor: backgroundColors.shift(),
        });
      });

      this.setState({
        statsUp: statsUp,
        statsDown: statsDown,
        statsUpFreq: statsUpFreq,
        statsDownFreq: statsDownFreq,
        statsUpDr: statsUpDr,
        statsDownDr: statsDownDr,
        statsDownError: statsDownError,
      });
    });
  }

  render() {
    if (this.props.gateway === undefined || this.state.statsDown === undefined || this.state.statsUp === undefined || this.state.statsUpFreq === undefined) {
      return(<div></div>);
    }

    const style = {
      height: 400,
    };

    const statsOptions = {
      animation: false,
      plugins: {
        legend: {
          display: false,
        },
      },
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
        },
        x: {
          type: "time",
        },
      },
    };

    const barOptions = {
      animation: false,
      plugins: {
        legend: {
          display: true,
        },
      },
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
        },
        x: {
          type: "time",
        },
      },
    };

    let position = [];
    if (typeof(this.props.gateway.location.latitude) !== "undefined" && typeof(this.props.gateway.location.longitude !== "undefined")) {
      position = [this.props.gateway.location.latitude, this.props.gateway.location.longitude]; 
    } else {
      position = [0,0];
    }

    let lastSeenAt = "Never";
    if (this.props.lastSeenAt !== null) {
      lastSeenAt = moment(this.props.lastSeenAt).format("lll");
    }

    return(
      <Grid container spacing={4}>
        <Grid item xs={6}>
          <DetailsCard gateway={this.props.gateway} lastSeenAt={lastSeenAt} />
        </Grid>
        <Grid item xs={6}>
          <Paper>
            <Map center={position} zoom={15} style={style} animate={true} scrollWheelZoom={false}>
              <MapTileLayer />
              <Marker position={position} />
            </Map>
          </Paper>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Received" />
            <CardContent className={this.props.classes.chart}>
              <Line height={75} options={statsOptions} data={this.state.statsDown} />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Transmitted" />
            <CardContent className={this.props.classes.chart}>
              <Line height={75} options={statsOptions} data={this.state.statsUp} />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Received / frequency" />
            <CardContent className={this.props.classes.chart}>
              <Heatmap data={this.state.statsUpFreq} fromColor="rgb(227, 242, 253)" toColor="rgb(33, 150, 243, 1)" />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Transmitted / frequency" />
            <CardContent className={this.props.classes.chart}>
              <Heatmap data={this.state.statsDownFreq} fromColor="rgb(227, 242, 253)" toColor="rgb(33, 150, 243, 1)" />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Received / DR" />
            <CardContent className={this.props.classes.chart}>
              <Heatmap data={this.state.statsUpDr} fromColor="rgb(227, 242, 253)" toColor="rgb(33, 150, 243, 1)" />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Transmitted / DR" />
            <CardContent className={this.props.classes.chart}>
              <Heatmap data={this.state.statsDownDr} fromColor="rgb(227, 242, 253)" toColor="rgb(33, 150, 243, 1)" />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Transmission / Ack status" />
            <CardContent className={this.props.classes.chart}>
              <Bar data={this.state.statsDownError} options={barOptions} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(GatewayDetails);
