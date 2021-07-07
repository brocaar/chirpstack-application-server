import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Avatar from "@material-ui/core/Avatar";
import Chip from "@material-ui/core/Chip";
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
import AlertCircleOutline from "mdi-material-ui/AlertCircleOutline";

import fileDownload from "js-file-download";
import moment from "moment";

import LoRaWANFrameLog from "../../components/LoRaWANFrameLog";
import DeviceStore from "../../stores/DeviceStore";
import theme from "../../theme";


const styles = {
  buttons: {
    textAlign: "right",
  },
  button: {
    marginLeft: 2 * theme.spacing(1),
  },
  icon: {
    marginRight: theme.spacing(1),
  },
  center: {
    textAlign: "center",
  },
  progress: {
    marginTop: 4 * theme.spacing(1),
  },
};


class DeviceFrames extends Component {
  constructor() {
    super();

    this.state = {
      paused: false,
      connected: false,
      frames: [],
      dialogOpen: false,
    };

    this.setConnected = this.setConnected.bind(this);
    this.onFrame = this.onFrame.bind(this);
    this.onDownload = this.onDownload.bind(this);
    this.togglePause = this.togglePause.bind(this);
    this.onClear = this.onClear.bind(this);
    this.toggleHelpDialog = this.toggleHelpDialog.bind(this);
  }

  componentDidMount() {
    const conn = DeviceStore.getFrameLogsConnection(this.props.match.params.devEUI, this.onFrame);
    this.setState({
      wsConn: conn,
    });

    DeviceStore.on("ws.status.change", this.setConnected);
    this.setConnected();
  }

  componentWillUnmount() {
    this.state.wsConn.close();

    DeviceStore.removeListener("ws.status.change", this.setConnected);
  }

  onDownload() {
    const dl = this.state.frames.map((frame, i) => {
      return {
        rxInfo: frame.rxInfo,
        txInfo: frame.txInfo,
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

  setConnected() {
    this.setState({
      connected: DeviceStore.getWSFramesStatus(),
    });
  }

  onFrame(frame) {
    if (this.state.paused) {
      return;
    }

    let frames = this.state.frames;
    let f = {};
    const now = new Date();

    if (frame.uplinkFrame !== undefined) {
      f = {
        id: now.getTime(),
        publishedAt: frame.uplinkFrame.publishedAt,
        rxInfo: frame.uplinkFrame.rxInfo,
        txInfo: frame.uplinkFrame.txInfo,
        phyPayload: JSON.parse(frame.uplinkFrame.phyPayloadJSON),
      };
    }

    if (frame.downlinkFrame !== undefined) {
      delete frame.downlinkFrame.txInfo['gatewayID'];

      f = {
        id: now.getTime(),
        publishedAt: frame.downlinkFrame.publishedAt,
        gatewayID: frame.downlinkFrame.gatewayID,
        txInfo: frame.downlinkFrame.txInfo,
        phyPayload: JSON.parse(frame.downlinkFrame.phyPayloadJSON),
      };
    }

    if (frames.length === 0 || moment(f.publishedAt).isAfter(moment(frames[0].publishedAt))) {
      frames.unshift(f);
      this.setState({
        frames: frames,
      });
    }
  }

  render() {
    const frames = this.state.frames.map((frame, i) => <LoRaWANFrameLog key={frame.id} frame={frame} />);

    return(
      <Grid container spacing={4}>
        <Dialog
          open={this.state.dialogOpen}
          onClose={this.toggleHelpDialog}
          aria-labelledby="help-dialog-title"
          aria-describedby="help-dialog-description"
        >
          <DialogTitle id="help-dialog-title">Help</DialogTitle>
          <DialogContent>
            <DialogContentText id="help-dialog-description">
              The frames below are the raw (and encrypted) LoRaWAN PHYPayload frames as seen by the gateway(s). This data is intedend for debugging only.
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={this.toggleHelpDialog} color="primary">Close</Button>
          </DialogActions>
        </Dialog>

        <Grid item xs={12} className={this.props.classes.buttons}>
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
          {!this.state.connected && <div className={this.props.classes.center}>
            <Chip
              color="secondary"
              label="Not connected to Websocket API"
              avatar={<Avatar><AlertCircleOutline /></Avatar>}
            />
          </div>}
          {(this.state.connected && frames.length === 0 && !this.state.paused) && <div className={this.props.classes.center}><CircularProgress className={this.props.classes.progress} /></div>}
          {frames.length > 0 && frames}
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(DeviceFrames);
