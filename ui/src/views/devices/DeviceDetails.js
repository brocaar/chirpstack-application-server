import React, { Component } from "react";
import { withRouter } from "react-router-dom";

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

import Refresh from "mdi-material-ui/Refresh";
import Delete from "mdi-material-ui/Delete";

import moment from "moment";

import TableCellLink from "../../components/TableCellLink";
import DeviceQueueItemForm from "./DeviceQueueItemForm";
import DeviceQueueStore from "../../stores/DeviceQueueStore";
import DeviceStore from "../../stores/DeviceStore";


class DetailsCard extends Component {
  render() {
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
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  }
}


class StatusCard extends Component {
  render() {
    let lastSeenAt = "never";

    if (this.props.device.lastSeenAt !== null) {
      lastSeenAt = moment(this.props.device.lastSeenAt).format("lll");
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
    };
  }

  componentDidMount() {
    this.setDeviceActivation();
  }

  componentDidUpdate(prevProps) {
    if (prevProps.device !== this.props.device) {
      this.setDeviceActivation();
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

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={6}>
          <DetailsCard device={this.props.device} deviceProfile={this.props.deviceProfile} match={this.props.match} />
        </Grid>
        <Grid item xs={6}>
          <StatusCard device={this.props.device} />
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

export default DeviceDetails;
