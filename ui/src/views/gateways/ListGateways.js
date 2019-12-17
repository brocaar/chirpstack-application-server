import React, { Component } from "react";
import { Route, Switch, Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Paper from '@material-ui/core/Paper';
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';

import Plus from "mdi-material-ui/Plus";

import moment from "moment";
import { Bar } from "react-chartjs-2";
import { Map, Marker, Popup } from 'react-leaflet';
import MarkerClusterGroup from "react-leaflet-markercluster";
import L from "leaflet";
import "leaflet.awesome-markers";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TableCellLink from "../../components/TableCellLink";
import TitleBarButton from "../../components/TitleBarButton";
import DataTable from "../../components/DataTable";
import GatewayAdmin from "../../components/GatewayAdmin";
import GatewayStore from "../../stores/GatewayStore";
import MapTileLayer from "../../components/MapTileLayer";

import theme from "../../theme";


class GatewayRow extends Component {
  constructor() {
    super();

    this.state = {};
  }

  componentDidMount() {
    const start = moment().subtract(29, "days").toISOString();
    const end = moment().toISOString();

    GatewayStore.getStats(this.props.gateway.id, start, end, resp => {
      let stats = {
        labels: [],
        datasets: [
          {
            data: [],
            fillColor: "rgba(33, 150, 243, 1)",
          },
        ],
      };

      for (const row of resp.result) {
        stats.labels.push(row.timestamp);
        stats.datasets[0].data.push(row.rxPacketsReceivedOK + row.txPacketsEmitted);
      }

      this.setState({
        stats: stats,
      });
    });
  }

  render() {
    const options = {
      scales: {
        xAxes: [{display: false}],
        yAxes: [{display: false}],
      },
      tooltips: {
        enabled: false,
      },
      legend: {
        display: false,
      },
      responsive: false,
      animation: {
        duration: 0,
      },
    };

    return(
      <TableRow>
          <TableCellLink to={`/organizations/${this.props.gateway.organizationID}/gateways/${this.props.gateway.id}`}>{this.props.gateway.name}</TableCellLink>
          <TableCell>{this.props.gateway.id}</TableCell>
          <TableCell>
            {this.state.stats && <Bar
              width={380}
              height={23}
              data={this.state.stats}
              options={options}
            />}
          </TableCell>
      </TableRow>
    );
  }
}


const styles = {
  chart: {
    width: 380,
  },
  tabs: {
    borderBottom: "1px solid " + theme.palette.divider,
    height: "48px",
    overflow: "visible",
  },
};



class ListGatewaysTable extends Component {
  constructor() {
    super();
    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    GatewayStore.list("", this.props.organizationID, limit, offset, callbackFunc);
  }

  getRow(obj) {
    return(
      <GatewayRow key={obj.id} gateway={obj} />
    );
  }

  render() {
    return(
      <DataTable
        header={
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>Gateway ID</TableCell>
            <TableCell className={this.props.classes.chart}>Gateway activity (30d)</TableCell>
          </TableRow>
        }
        getPage={this.getPage}
        getRow={this.getRow}
      />
    );
  }
}
ListGatewaysTable = withStyles(styles)(ListGatewaysTable);


class ListGatewaysMap extends Component {
  constructor() {
    super();

    this.state = {
      items: null,
    };
  }

  componentDidMount() {
    GatewayStore.list("", this.props.organizationID, 9999, 0, resp => {
      this.setState({
        items: resp.result,
      });
    });
  }

  render() {
    if (this.state.items === null || this.state.items.length === 0) {
      return null;
    }

    const style = {
      height: 800,
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
            <Link to={`/organizations/${this.props.organizationID}/gateways/${item.id}`}>{item.name}</Link><br />
            {item.id}<br /><br />
            {lastSeen}
          </Popup>
        </Marker>
      );
    }

    return(
      <Paper>
        <Map bounds={bounds} maxZoom={19} style={style} animate={true} scrollWheelZoom={false}>
          <MapTileLayer />
          <MarkerClusterGroup>
            {markers}
          </MarkerClusterGroup>
        </Map>
      </Paper>
    );
  }
}


class ListGateways extends Component {
  constructor() {
    super();

    this.state = {
      tab: 0,
    };
  }

  componentDidMount() {
    this.locationToTab();
  }

  onChangeTab = (e, v) => {
    this.setState({
      tab: v,
    });
  };

  locationToTab = () => {
    let tab = 0;

    if (window.location.href.endsWith("/map")) {
      tab = 1;
    }

    this.setState({
      tab: tab,
    });
  };

  render() {
    return(
      <Grid container spacing={4}>
        <TitleBar
          buttons={
            <GatewayAdmin organizationID={this.props.match.params.organizationID}>
              <TitleBarButton
                key={1}
                label="Create"
                icon={<Plus />}
                to={`/organizations/${this.props.match.params.organizationID}/gateways/create`}
              />
            </GatewayAdmin>
          }
        >
          <TitleBarTitle title="Gateways" />
        </TitleBar>

        <Grid item xs={12}>
          <Tabs
            value={this.state.tab}
            onChange={this.onChangeTab}
            indicatorColor="primary"
            className={this.props.classes.tabs}
          >
            <Tab label="List" component={Link} to={`/organizations/${this.props.match.params.organizationID}/gateways`} />
            <Tab label="Map" component={Link} to={`/organizations/${this.props.match.params.organizationID}/gateways/map`} />
          </Tabs>
        </Grid>

        <Grid item xs={12}>
          <Switch>
            <Route exact path={this.props.match.path} render={props => <ListGatewaysTable {...props} organizationID={this.props.match.params.organizationID} />} />
            <Route exact path={`${this.props.match.path}/map`} render={props => <ListGatewaysMap {...props} organizationID={this.props.match.params.organizationID} />} />
          </Switch>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(ListGateways);
