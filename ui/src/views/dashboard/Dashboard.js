import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";

import moment from "moment";
import { Map, Marker, Popup } from 'react-leaflet';
import MarkerClusterGroup from "react-leaflet-markercluster";
import L from "leaflet";
import { Doughnut } from 'react-chartjs-2';
import "leaflet.awesome-markers";

import MapTileLayer from "../../components/MapTileLayer";
import GatewayStore from "../../stores/GatewayStore";
import InternalStore from "../../stores/InternalStore";
import theme from "../../theme";


const styles = {
  doughtnutChart: {
    maxWidth: "350px",
    padding: 0,
    margin: "auto",
    display: "block",
  },
};


class ListGatewaysMap extends Component {
  constructor() {
    super();

    this.state = {
      items: null,
    };
  }

  componentDidMount() {
    this.loadData();
  }

  componentDidUpdate(prevProps) {
    if (prevProps === this.props) {
      return;
    }

    this.loadData();
  }

  loadData = () => {
    GatewayStore.list("", 0, 9999, 0, resp => {
      this.setState({
        items: resp.result,
      });
    });
  }

  render() {
    if (this.state.items === null || this.state.items.length === 0) {
      return(
        <Card>
          <CardHeader title="Gateways" />
            <CardContent>
              No data available.
            </CardContent>
        </Card>
      );
    }

    const style = {
      height: 600,
    };


    let bounds = [];
    let markers = [];

    const greenMarker = L.AwesomeMarkers.icon({
      icon: "wifi",
      prefix: "fa",
      markerColor: "green",
    });

    const grayMarker = L.AwesomeMarkers.icon({
      icon: "wifi",
      prefix: "fa",
      markerColor: "gray",
    });

    const redMarker = L.AwesomeMarkers.icon({
      icon: "wifi",
      prefix: "fa",
      markerColor: "red",
    });

    for (const item of this.state.items) {
      const position = [item.location.latitude, item.location.longitude];

      bounds.push(position);

      let marker = greenMarker;
      let lastSeen = "";

      if (item.lastSeenAt === undefined || item.lastSeenAt === null) {
        marker = grayMarker;
        lastSeen = "Never seen online";
      } else {
        const ts = moment(item.lastSeenAt);
        if (ts.isBefore(moment().subtract(5, 'minutes'))) {
          marker = redMarker;
        }

        lastSeen = ts.fromNow();
      }

      markers.push(
        <Marker position={position} key={`gw-${item.id}`} icon={marker}>
          <Popup>
            <Link to={`/organizations/${item.organizationID}/gateways/${item.id}`}>{item.name}</Link><br />
            {item.id}<br /><br />
            {lastSeen}
          </Popup>
        </Marker>
      );
    }

    return(
      <Card>
        <CardHeader title="Gateways" />
          <CardContent>
            <Map bounds={bounds} maxZoom={19} style={style} animate={true} scrollWheelZoom={false}>
              <MapTileLayer />
              <MarkerClusterGroup>
                {markers}
              </MarkerClusterGroup>
            </Map>
          </CardContent>
      </Card>
    );
  }
}


class DevicesActiveInactive extends Component {
  render() {
    let data = null;

    if (this.props.summary !== null && (this.props.summary.activeCount !== 0 || this.props.summary.inactiveCount !== 0)) {
      data = {
        labels: ["Never seen", "Inactive", "Active"],
        datasets: [
          {
            data: [this.props.summary.neverSeenCount, this.props.summary.inactiveCount, this.props.summary.activeCount],
            backgroundColor: [
              theme.palette.warning.main,
              theme.palette.error.main,
              theme.palette.success.main,
            ],
          },
        ],
      };
    }

    const options = {
      animation: false,
    };

    return(
      <Card>
        <CardHeader title="Active devices" />
        <CardContent>
          {data && <Doughnut data={data} options={options} className={this.props.classes.doughtnutChart} />}
          {!data && <div>No data available.</div>}
        </CardContent>
      </Card>
    );
  }
}

DevicesActiveInactive = withStyles(styles)(DevicesActiveInactive);


class GatewaysActiveInactive extends Component {
  render() {
    let data = null;

    if (this.props.summary !== null && (this.props.summary.activeCount !== 0 || this.props.summary.inactiveCount !== 0)) {
      data = {
        labels: ["Never seen", "Inactive", "Active"],
        datasets: [
          {
            data: [this.props.summary.neverSeenCount, this.props.summary.inactiveCount, this.props.summary.activeCount],
            backgroundColor: [
              theme.palette.warning.main,
              theme.palette.error.main,
              theme.palette.success.main,
            ],
          },
        ],
      };
    }

    const options = {
      animation: false,
    };


    return(
      <Card>
        <CardHeader title="Active gateways" />
        <CardContent>
          {data && <Doughnut data={data} options={options} className={this.props.classes.doughtnutChart} />}
          {!data && <div>No data available.</div>}
        </CardContent>
      </Card>
    );
  }
}

GatewaysActiveInactive = withStyles(styles)(GatewaysActiveInactive);


class DevicesDataRates extends Component {
  getColor = (dr) => {
    return ['#ff5722', '#ff9800', '#ffc107', '#ffeb3b', '#cddc39', '#8bc34a', '#4caf50', '#009688', '#00bcd4', '#03a9f4', '#2196f3', '#3f51b5', '#673ab7', '#9c27b0', '#e91e63'][dr];
  }

  render() {
    let data = null;

    if (this.props.summary !== null && Object.keys(this.props.summary.drCount).length !== 0) {
      data = {
        labels: [],
        datasets: [
          {
            data: [],
            backgroundColor: [],
          },
        ],
      };

      for (let dr in this.props.summary.drCount) {
        data.labels.push(`DR${dr}`);
        data.datasets[0].data.push(this.props.summary.drCount[dr]);
        data.datasets[0].backgroundColor.push(this.getColor(dr));
      }
    }
    
    const options = {
      animation: false,
    };

    return(
      <Card>
        <CardHeader title="Device data-rate usage" />
        <CardContent>
          {data && <Doughnut data={data} options={options} className={this.props.classes.doughtnutChart} />}
          {!data && <div>No data available.</div>}
        </CardContent>
      </Card>
    );
  }
}

DevicesDataRates = withStyles(styles)(DevicesDataRates);


class Dashboard extends Component {
  constructor() {
    super();

    this.state = {
      devicesSummary: null,
      gatewaysSummary: null,
    };
  }

  componentDidMount() {
    InternalStore.getDevicesSummary(0, resp => {
      this.setState({
        devicesSummary: resp,
      });
    });

    InternalStore.getGatewaysSummary(0, resp => {
      this.setState({
        gatewaysSummary: resp,
      });
    });
  }

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Dashboard" />
        </TitleBar>

        <Grid item xs={12}>
          <Grid container spacing={4}>
            <Grid item xs={4}>
              <DevicesActiveInactive summary={this.state.devicesSummary} />
            </Grid>

            <Grid item xs={4}>
              <GatewaysActiveInactive summary={this.state.gatewaysSummary} />
            </Grid>

            <Grid item xs={4}>
              <DevicesDataRates summary={this.state.devicesSummary} />
            </Grid>

            <Grid item xs={12}>
              <ListGatewaysMap />
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    );
  }
}


export default Dashboard;
