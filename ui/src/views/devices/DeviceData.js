import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import Button from '@material-ui/core/Button';
import CircularProgress from '@material-ui/core/CircularProgress';
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import Typography from '@material-ui/core/Typography';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

import Play from "mdi-material-ui/Play";
import Pause from "mdi-material-ui/Pause";
import Download from "mdi-material-ui/Download";
import Delete from "mdi-material-ui/Delete";
import ChevronDown from "mdi-material-ui/ChevronDown";
import HelpCircleOutline from "mdi-material-ui/HelpCircleOutline";

import moment from "moment";
import fileDownload from "js-file-download";

import DeviceStore from "../../stores/DeviceStore";
import theme from "../../theme";
import JSONTree from "../../components/JSONTree";


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
  headerColumn: {
    marginRight: 6 * theme.spacing.unit,
  },
  headerColumnFixedSmall: {
    width: 145,
  },
  headerColumnFixedWide: {
    width: 175,
  },
  treeStyle: {
    paddingTop: 0,
    paddingBottom: 0,
    fontSize: 12,
    lineHeight: 1.5,
  },
};


class DeviceDataItem extends Component {
  render() {
    const receivedAt = moment(this.props.data.receivedAt).format("LTS");
    
    return(
      <ExpansionPanel>
        <ExpansionPanelSummary expandIcon={<ChevronDown />}>
          <div className={this.props.classes.headerColumnFixedSmall}><Typography variant="body2">{receivedAt}</Typography></div>
          <div className={this.props.classes.headerColumnFixedSmall}><Typography variant="body2">{this.props.data.type}</Typography></div>
        </ExpansionPanelSummary>
        <ExpansionPanelDetails>
          <Grid container spacing={24}>
            <Grid item xs className={this.props.classes.treeStyle}>
              <JSONTree data={this.props.data.payload} />
            </Grid>
          </Grid>
        </ExpansionPanelDetails>
      </ExpansionPanel>
    );
  }
}

DeviceDataItem = withStyles(styles)(DeviceDataItem);


class DeviceData extends Component {
  constructor() {
    super();

    this.state = {
      paused: false,
      data: [],
      dialogOpen: false,
    };

    this.onData = this.onData.bind(this);
    this.onDownload = this.onDownload.bind(this);
    this.togglePause = this.togglePause.bind(this);
    this.onClear = this.onClear.bind(this);
    this.toggleHelpDialog = this.toggleHelpDialog.bind(this);
  }

  componentDidMount() {
    const conn = DeviceStore.getDataLogsConnection(this.props.match.params.devEUI, this.onData);
    this.setState({
      wsConn: conn,
    });
  }

  componentWillUnmount() {
    this.state.wsConn.close();
  }

  onDownload() {
    const dl = this.state.data.map((data, i) => {
      return {
        type: data.type,
        payload: data.payload,
      };
    });

    fileDownload(JSON.stringify(dl, null, 4), `device-${this.props.match.params.devEUI}.json`);
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
      data: [],
    });
  }

  onData(d) {
    if (this.state.paused) {
      return;
    }

    let data = this.state.data;
    const now = new Date();

    data.unshift({
      id: now.getTime(),
      receivedAt: now,
      type: d.type,
      payload: JSON.parse(d.payloadJSON),
    });

    this.setState({
      data: data,
    });
  }

  render() {
    const data = this.state.data.map((d, i) => <DeviceDataItem key={d.id} data={d} />);

    return(
      <Grid container spacing={24}>
        <Dialog
          open={this.state.dialogOpen}
          onClose={this.toggleHelpDialog}
          aria-labelledby="help-dialog-title"
          aria-describedby="help-dialog-description"
        >
          <DialogTitle id="help-dialog-title">Help</DialogTitle>
          <DialogContent>
            <DialogContentText id="help-dialog-description">
              These are the events as published to the application. Please refer to data integrations (documentation) for more information on integrating this with your application.
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
          {data.length > 0 && data}
          {(data.length === 0 && !this.state.paused) && <div className={this.props.classes.center}><CircularProgress className={this.props.classes.progress} /></div>}
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(DeviceData);
