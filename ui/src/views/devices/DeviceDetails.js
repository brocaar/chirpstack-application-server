import React, { Component } from "react";
import { withRouter, Link } from "react-router-dom";

import Grid from "@material-ui/core/Grid";
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import Table from "@material-ui/core/Table";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";
import Button from '@material-ui/core/Button';
import { withStyles } from '@material-ui/core/styles';

import Refresh from "mdi-material-ui/Refresh";
import Delete from "mdi-material-ui/Delete";

import moment from "moment";
import { Line, Bar } from "react-chartjs-2";

import Heatmap from "../../components/Heatmap";
import TableCellLink from "../../components/TableCellLink";
import DeviceQueueItemForm from "./DeviceQueueItemForm";
import DeviceQueueStore from "../../stores/DeviceQueueStore";
import DeviceStore from "../../stores/DeviceStore";
import MulticastGroupStore from "../../stores/MulticastGroupStore";

import theme from "../../theme";


const styles = {
  link: {
    textDecoration: "none",
    color: theme.palette.primary.main,
    cursor: "pointer",
  },
  chart: {
    height: 300,
  },
};



class DetailsCard extends Component {
  render() {
    const multicastGroups = this.props.multicastGroups.map((mg, i) => <div><Link className={this.props.classes.link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/multicast-groups/${mg.id}`}>{ mg.name }</Link></div>);

    return(
      <Card>
        <CardHeader title="Details" />
        <CardContent>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>{this.props.device.device.name}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Description</TableCell>
                <TableCell>{this.props.device.device.description}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Device-profile</TableCell>
                <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/device-profiles/${this.props.deviceProfile.deviceProfile.id}`}>{this.props.deviceProfile.deviceProfile.name}</TableCellLink>
              </TableRow>
              <TableRow>
                <TableCell>Multicast groups</TableCell>
                <TableCell>{multicastGroups}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  }
}

DetailsCard = withStyles(styles)(DetailsCard);


class StatusCard extends Component {
  render() {
    let lastSeenAt = "never";
    let state = "enabled";

    if (this.props.device.lastSeenAt !== null) {
      lastSeenAt = moment(this.props.device.lastSeenAt).format("lll");
    }

    if (this.props.device.device.isDisabled === true) {
      state = "disabled";
    }

    return(
      <Card>
        <CardHeader title="Status" />
        <CardContent>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell>Last seen at</TableCell>
                <TableCell>{lastSeenAt}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>State</TableCell>
                <TableCell>{state}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  }
}

class EnqueueCard extends Component {
  constructor() {
    super();

    this.state = {
      object: {},
    };
  }

  onSubmit = (queueItem) => {
    let qi = queueItem;
    qi.devEUI = this.props.match.params.devEUI;

    DeviceQueueStore.enqueue(qi, resp => {
      this.setState({
        object: {},
      });
    });
  }

  render() {
    return(
      <Card>
        <CardHeader title="Enqueue downlink payload" />
        <CardContent>
          <DeviceQueueItemForm
            submitLabel="Enqueue payload"
            onSubmit={this.onSubmit}
            object={this.state.object}
          />
        </CardContent>
      </Card>
    );
  }
}

EnqueueCard = withRouter(EnqueueCard);


class QueueCardRow extends Component {
  render() {
    let confirmed = "no";
    if (this.props.item.confirmed) {
      confirmed = "yes";
    }

    return(
      <TableRow>
        <TableCell>{this.props.item.fCnt}</TableCell>
        <TableCell>{this.props.item.fPort}</TableCell>
        <TableCell>{confirmed}</TableCell>
        <TableCell>{this.props.item.data}</TableCell>
      </TableRow>
    );
  }
}


class QueueCard extends Component {
  constructor() {
    super();

    this.state = {
      deviceQueueItems: [],
    };
  }

  componentDidMount() {
    this.getQueue();

    DeviceQueueStore.on("enqueue", this.getQueue);
  }

  componentWillUnmount() {
    DeviceQueueStore.removeListener("enqueue", this.getQueue);
  }

  getQueue = () => {
    DeviceQueueStore.list(this.props.match.params.devEUI, resp => {
      this.setState({
        deviceQueueItems: resp.deviceQueueItems,
      });
    });
  }

  flushQueue = () => {
    if (window.confirm("Are you sure you want to flush the device queue?")) {
      DeviceQueueStore.flush(this.props.match.params.devEUI, resp => {
        this.getQueue();
      });
    }
  }

  render() {
    const rows = this.state.deviceQueueItems.map((item, i) => <QueueCardRow key={i} item={item}/>);

    return(
      <Card>
        <CardHeader title="Downlink queue" action={
          <div>
            <Button onClick={this.getQueue}><Refresh /></Button>
            <Button onClick={this.flushQueue} color="secondary"><Delete /></Button>
          </div>
        } />
        <CardContent>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>FCnt</TableCell>
                <TableCell>FPort</TableCell>
                <TableCell>Confirmed</TableCell>
                <TableCell>Base64 encoded payload</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  }
}

QueueCard = withRouter(QueueCard);


class DeviceDetails extends Component {
  constructor() {
    super();
    this.state = {
      activated: false,
      multicastGroups: [],
    };
  }

  componentDidMount() {
    this.setDeviceActivation();
    this.setMulticastGroups();
    this.loadStats();
  }

  componentDidUpdate(prevProps) {
    if (prevProps.device !== this.props.device) {
      this.setDeviceActivation();
      this.setMulticastGroups();
    }
  }

  setDeviceActivation = () => {
    if (this.props.device === undefined) {
      return;
    }

    DeviceStore.getActivation(this.props.device.device.devEUI, resp => {
      if (resp === null) {
        this.setState({
          activated: false,
        });
      } else {
        this.setState({
          activated: true,
        });
      }
    });
  };

  setMulticastGroups = () => {
    MulticastGroupStore.list("", 0, "", this.props.device.device.devEUI, 999, 0, resp => {
      this.setState({
        multicastGroups: resp.result,
      });
    });
  }

  loadStats = () => {
    const end = moment().toISOString();
    const start = moment().subtract(30, "days").toISOString();

    DeviceStore.getStats(this.props.match.params.devEUI, start, end, resp => {
      let statsUp = {
        labels: [],
        datasets: [
          {
            label: "uplink",
            borderColor: "rgba(33, 150, 243, 1)",
            backgroundColor: "rgba(0, 0, 0, 0)",
            lineTension: 0,
            pointBackgroundColor: "rgba(33, 150, 243, 1)",
            data: [],
          },
        ],
      };
      
      let statsUpFreq = [];

      let statsGwRssi = {
        labels: [],
        datasets: [
          {
            label: "rssi (reported by gateways)",
            borderColor: "rgba(33, 150, 243, 1)",
            backgroundColor: "rgba(0, 0, 0, 0)",
            lineTension: 0,
            pointBackgroundColor: "rgba(33, 150, 243, 1)",
            data: [],
          },
        ],
      };

      let statsGwSnr = {
        labels: [],
        datasets: [
          {
            label: "snr (reported by gateways)",
            borderColor: "rgba(33, 150, 243, 1)",
            backgroundColor: "rgba(0, 0, 0, 0)",
            lineTension: 0,
            pointBackgroundColor: "rgba(33, 150, 243, 1)",
            data: [],
          },
        ],
      };

      let statsErrors = {
        labels: [],
        datasets: [],
      };
      let statsErrorsSet = {};

      let statsUpDr = {
        labels: [],
        datasets: [],
      };
      let statsUpDrSet = [];

      for (const row of resp.result) {
        statsUp.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));
        statsUp.datasets[0].data.push(row.rxPackets);

        statsGwRssi.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));
        statsGwSnr.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));

        if (row.rxPackets !== 0) {
          statsGwRssi.datasets[0].data.push(row.gwRssi);
          statsGwSnr.datasets[0].data.push(row.gwSnr);
        } else {
          statsGwRssi.datasets[0].data.push(null);
          statsGwSnr.datasets[0].data.push(null);
        }

        statsUpFreq.push({
          x: moment(row.timestamp).format("YYYY-MM-DD"),
          y: row.rxPacketsPerFrequency,
        });

        statsErrors.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));
        Object.entries(row.errors).forEach(([k, v]) => {
          if (statsErrorsSet[k] === undefined) {
            statsErrorsSet[k] = [];
          }

          // fill gaps with 0s
          for (let i = statsErrorsSet[k].length; i < statsErrors.labels.length - 1; i++) {
            statsErrorsSet[k].push(0);
          }

          statsErrorsSet[k].push(v);
        });

        statsUpDr.labels.push(moment(row.timestamp).format("YYYY-MM-DD"));
        Object.entries(row.rxPacketsPerDr).forEach(([k, v]) => {
          if (statsUpDrSet[k] === undefined) {
            statsUpDrSet[k] = [];
          }

          // fill gaps with 0s
          for (let i = statsUpDrSet[k].length; i < statsUpDr.labels.length - 1; i++) {
            statsUpDrSet[k].push(0);
          }

          statsUpDrSet[k].push(v);
        });
      }

      let backgroundColors = ['#8bc34a', '#ff5722', '#ff9800', '#ffc107', '#ffeb3b', '#cddc39', '#4caf50', '#009688', '#00bcd4', '#03a9f4', '#2196f3', '#3f51b5', '#673ab7', '#9c27b0', '#e91e63'];

      Object.entries(statsErrorsSet).forEach(([k, v]) => {
        statsErrors.datasets.push({
          label: k,
          data: v,
          backgroundColor: backgroundColors.shift(),
        });
      });

      backgroundColors = ['#8bc34a', '#ff5722', '#ff9800', '#ffc107', '#ffeb3b', '#cddc39', '#4caf50', '#009688', '#00bcd4', '#03a9f4', '#2196f3', '#3f51b5', '#673ab7', '#9c27b0', '#e91e63'];

      Object.entries(statsUpDrSet).forEach(([k, v]) => {
        statsUpDr.datasets.push({
          label: k,
          data: v,
          backgroundColor: backgroundColors.shift(),
        });
      });

      this.setState({
        statsUp: statsUp,
        statsGwRssi: statsGwRssi,
        statsGwSnr: statsGwSnr,
        statsUpFreq: statsUpFreq,
        statsErrors: statsErrors,
        statsUpDr: statsUpDr,
      });
    });
  }

  render() {
    if (this.state.statsUp === undefined) {
      return null;
    }

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

    return(
      <Grid container spacing={4}>
        <Grid item xs={6}>
          <DetailsCard multicastGroups={this.state.multicastGroups} device={this.props.device} deviceProfile={this.props.deviceProfile} match={this.props.match} />
        </Grid>
        <Grid item xs={6}>
          <StatusCard device={this.props.device} />
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Received" />
            <CardContent className={this.props.classes.chart}>
              <Line height={75} options={statsOptions} data={this.state.statsUp} />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Errors" />
            <CardContent className={this.props.classes.chart}>
              <Bar data={this.state.statsErrors} options={barOptions} />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="SNR" />
            <CardContent className={this.props.classes.chart}>
              <Line height={75} options={statsOptions} data={this.state.statsGwSnr} />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="RSSI" />
            <CardContent className={this.props.classes.chart}>
              <Line height={75} options={statsOptions} data={this.state.statsGwRssi} />
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
            <CardHeader title="Received / DR" />
            <CardContent className={this.props.classes.chart}>
              <Bar data={this.state.statsUpDr} options={barOptions} />
            </CardContent>
          </Card>
        </Grid>
        {this.state.activated && <Grid item xs={12}>
          <EnqueueCard />
        </Grid>}
        {this.state.activated &&<Grid item xs={12}>
          <QueueCard />
        </Grid>}
      </Grid>
    );
  }
}

export default withStyles(styles)(DeviceDetails);
