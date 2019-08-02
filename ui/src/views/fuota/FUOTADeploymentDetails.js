import React, { Component } from "react";

import Grid from '@material-ui/core/Grid';
import Card from "@material-ui/core/Card";
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import LinearProgress from '@material-ui/core/LinearProgress';
import Typography from "@material-ui/core/Typography";

import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';

import moment from "moment";

import FUOTADeploymentStore from "../../stores/FUOTADeploymentStore";


class FUOTADeploymentDetails extends Component {
  constructor() {
    super();

    this.state = {
        progress: 0,
        stepProgress: 0,
        lastReload: 0,
    };
  }

  componentDidMount() {
    this.timer = setInterval(this.progress, 1000);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  componentDidUpdate(prevProps) {
    if (this.props === prevProps) {
      return;
    }

    this.progress();
  }

  progress = () => {
    const start = moment(this.props.fuotaDeployment.updatedAt).unix();
    const nextStepAfter = moment(this.props.fuotaDeployment.fuotaDeployment.nextStepAfter).unix();
    const now = moment().unix();

    const span = nextStepAfter - start;
    const stepProgress = now - start;
    const progressScaled = stepProgress / span * 100;

    if (progressScaled > 100) {
      if (moment().unix() - this.state.lastReload > 5) {
        this.setState({
          lastReload: moment().unix(),
        }, FUOTADeploymentStore.emitReload());
      }
    }

    const states = 8;
    let state = 0;

    switch(this.props.fuotaDeployment.fuotaDeployment.state) {
      case "MC_CREATE":
        state = 0;
        break;
      case "MC_SETUP":
        state = 1;
        break;
      case "FRAG_SESS_SETUP":
        state = 2;
        break;
      case "MC_SESS_C_SETUP":
        state = 3;
        break;
      case "ENQUEUE":
        state = 4;
        break;
      case "STATUS_REQUEST":
        state = 5;
        break;
      case "SET_DEVICE_STATUS":
        state = 6;
        break;
      case "CLEANUP":
        state = 7;
        break;
      case "DONE":
        state = 8;
        break;
      default:
        state = 0;
        break;
    }

    this.setState({
      stepProgress: progressScaled,
      progress: state / states * 100,
    });
  }

  render() {
    let multicastTimeout = 0;
    if (this.props.fuotaDeployment.fuotaDeployment.groupType === "CLASS_C") {
      multicastTimeout = (1 << this.props.fuotaDeployment.fuotaDeployment.multicastTimeout);
    }

    const createdAt = moment(this.props.fuotaDeployment.createdAt).format('lll');
    const updatedAt = moment(this.props.fuotaDeployment.updatedAt).format('lll');
    const nextStepAfter = moment(this.props.fuotaDeployment.fuotaDeployment.nextStepAfter).format('lll');

    return(
      <Grid container spacing={4}>
        <Grid item xs={6}>
          <Card>
            <CardHeader title="Details" />
            <CardContent>
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell>Job name</TableCell>
                    <TableCell>{this.props.fuotaDeployment.fuotaDeployment.name}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Redundant frames</TableCell>
                    <TableCell>{this.props.fuotaDeployment.fuotaDeployment.redundancy}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Unicast timeout</TableCell>
                    <TableCell>{this.props.fuotaDeployment.fuotaDeployment.unicastTimeout}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Data-rate</TableCell>
                    <TableCell>{this.props.fuotaDeployment.fuotaDeployment.dr}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Frequency</TableCell>
                    <TableCell>{this.props.fuotaDeployment.fuotaDeployment.frequency}Hz</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Multicast-group type</TableCell>
                    <TableCell>{this.props.fuotaDeployment.fuotaDeployment.groupType}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Multicast timeout</TableCell>
                    <TableCell>{multicastTimeout} seconds</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={6}>
          <Card>
            <CardHeader title="Status" />
            <CardContent>
              <Table>
                <TableBody>
                  <TableRow>
                    <TableCell>Created at</TableCell>
                    <TableCell>{createdAt}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Updated at</TableCell>
                    <TableCell>{updatedAt}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>State</TableCell>
                    <TableCell>{this.props.fuotaDeployment.fuotaDeployment.state}</TableCell>
                  </TableRow>
                  {this.props.fuotaDeployment.fuotaDeployment.state !== "DONE" && <TableRow>
                    <TableCell>Next step at</TableCell>
                    <TableCell>{nextStepAfter}</TableCell>
                  </TableRow>}
                </TableBody>
              </Table>
            </CardContent>
            {this.props.fuotaDeployment.fuotaDeployment.state !== "DONE" && <CardContent>
              <Typography variant="subtitle2" gutterBottom>
                Job progress:
              </Typography>
              <LinearProgress variant="determinate" value={this.state.progress} />
            </CardContent>}
            {this.props.fuotaDeployment.fuotaDeployment.state !== "DONE" && <CardContent>
              <Typography variant="subtitle2" gutterBottom>
                State progress:
              </Typography>
                <LinearProgress variant="determinate" value={this.state.stepProgress} />
            </CardContent>}
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default FUOTADeploymentDetails;

