import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import Button from '@material-ui/core/Button';
import CircularProgress from '@material-ui/core/CircularProgress';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

import Play from "mdi-material-ui/Play";
import Pause from "mdi-material-ui/Pause";
import Download from "mdi-material-ui/Download";
import Delete from "mdi-material-ui/Delete";
import HelpCircleOutline from "mdi-material-ui/HelpCircleOutline";

import fileDownload from "js-file-download";

import LoRaWANFrameLog from "../../components/LoRaWANFrameLog";
import GatewayStore from "../../stores/GatewayStore";
import theme from "../../theme";


const styles = {
  buttons: {
    textAlign: "right",
  },
  button: {
    marginLeft: 2 * theme.spacing.unit,
  },
  icon: {
    marginRight: theme.spacing.unit,
  },
  center: {
    textAlign: "center",
  },
  progress: {
    marginTop: 4 * theme.spacing.unit,
  },
};


class GatewayFrames extends Component {
  constructor() {
    super();

    this.state = {
      paused: false,
      frames: [],
      dialogOpen: false,
    };

    this.onFrame = this.onFrame.bind(this);
    this.onDownload = this.onDownload.bind(this);
    this.togglePause = this.togglePause.bind(this);
    this.onClear = this.onClear.bind(this);
    this.toggleHelpDialog = this.toggleHelpDialog.bind(this);
  }

  componentDidMount() {
    const conn = GatewayStore.getFrameLogsConnection(this.props.gateway.id, () => {}, () => {}, this.onFrame);
    this.setState({
      wsConn: conn,
    });
  }

  componentWillUnmount() {
    this.state.wsConn.close();
  }

  onDownload() {
    const dl = this.state.frames.map((frame, i) => {
      return {
        uplinkMetaData: frame.uplinkMetaData,
        downlinkMetaData: frame.downlinkMetaData,
        phyPayload: frame.phyPayload,
      };
    });

    fileDownload(JSON.stringify(dl, null, 4), `gateway-${this.props.match.params.gatewayID}.json`);
  }

  togglePause() {
    this.setState({
      paused: !this.state.paused,
    });
  }

  toggleHelpDialog() {
    this.setState({
      dialogOpen: !this.state.dialogOpen,
    });
  }

  onClear() {
    this.setState({
      frames: [],
    });
  }

  onFrame(frame) {
    if (this.state.paused) {
      return;
    }

    let frames = this.state.frames;
    const now = new Date();

    if (frame.uplinkFrame !== undefined) {
      frames.unshift({
        id: now.getTime(),
        receivedAt: now,
        uplinkMetaData: {
          rxInfo: frame.uplinkFrame.rxInfo,
          txInfo: frame.uplinkFrame.txInfo,
        },
        phyPayload: JSON.parse(frame.uplinkFrame.phyPayloadJSON),
      });
    }

    if (frame.downlinkFrame !== undefined) {
      frames.unshift({
        id: now.getTime(),
        receivedAt: now,
        downlinkMetaData: {
          txInfo: frame.downlinkFrame.txInfo,
        },
        phyPayload: JSON.parse(frame.downlinkFrame.phyPayloadJSON),
      });
    }

    console.log(frame);

    this.setState({
      frames: frames,
    });
  }

  render() {
    const frames = this.state.frames.map((frame, i) => <LoRaWANFrameLog key={frame.id} frame={frame} />);

    return(
      <Grid container spacing={24}>
        <Grid item xs={12} className={this.props.classes.buttons}>
          <Dialog
            open={this.state.dialogOpen}
            onClose={this.toggleHelpDialog}
            aria-labelledby="help-dialog-title"
            aria-describedby="help-dialog-description"
          >
            <DialogTitle id="help-dialog-title">Help</DialogTitle>
            <DialogContent>
              <DialogContentText id="help-dialog-description">
                The frames below are the raw (and encrypted) LoRaWAN PHYPayload frames as seen by the gateway(s). This data is intended for debugging only.
              </DialogContentText>
            </DialogContent>
            <DialogActions>
              <Button onClick={this.toggleHelpDialog} color="primary">Close</Button>
            </DialogActions>
          </Dialog>

          <Button variant="outlined" className={this.props.classes.button} onClick={this.toggleHelpDialog}>
            <HelpCircleOutline className={this.props.classes.icon} />
            Help
          </Button>
          {!this.state.paused && <Button variant="outlined" className={this.props.classes.button} onClick={this.togglePause}>
            <Pause className={this.props.classes.icon} />
            Pause
          </Button>}
          {this.state.paused && <Button variant="outlined" className={this.props.classes.button} onClick={this.togglePause}>
            <Play className={this.props.classes.icon} />
            Resume
          </Button>}
          <Button variant="outlined" className={this.props.classes.button} onClick={this.onDownload}>
            <Download className={this.props.classes.icon} />
            Download
          </Button>
          <Button variant="outlined" className={this.props.classes.button} color="secondary" onClick={this.onClear}>
            <Delete className={this.props.classes.icon} />
            Clear
          </Button>
        </Grid>
        <Grid item xs={12}>
          {frames.length > 0 && frames}
          {(frames.length === 0 && !this.state.paused) && <div className={this.props.classes.center}><CircularProgress className={this.props.classes.progress} /></div>}
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(GatewayFrames);
