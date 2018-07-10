import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import Typography from '@material-ui/core/Typography';

import ChevronDown from "mdi-material-ui/ChevronDown";

import moment from "moment";

import JSONTree from "./JSONTree";
import theme from "../theme";


const styles = {
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


class LoRaWANFrameLog extends Component {
  render() {
    let dir = "";
    let devID = "";

    if (this.props.frame.uplinkMetaData !== undefined) {
      dir = "UPLINK";
    } else {
      dir = "DOWNLINK";
    }

    const receivedAt = moment(this.props.frame.receivedAt).format("LTS");
    const mType = this.props.frame.phyPayload.mhdr.mType;

    if (this.props.frame.phyPayload.macPayload !== undefined) {
      if (this.props.frame.phyPayload.macPayload.devEUI !== undefined) {
        devID = this.props.frame.phyPayload.macPayload.devEUI;
      }

      if (this.props.frame.phyPayload.macPayload.fhdr !== undefined) {
        devID = this.props.frame.phyPayload.macPayload.fhdr.devAddr;
      }
    }

    return(
      <ExpansionPanel>
        <ExpansionPanelSummary expandIcon={<ChevronDown />}>
          <div className={this.props.classes.headerColumnFixedSmall}><Typography variant="body2">{dir}</Typography></div>
          <div className={this.props.classes.headerColumnFixedSmall}><Typography variant="body2">{receivedAt}</Typography></div>
          <div className={this.props.classes.headerColumnFixedWide}><Typography variant="body2">{mType}</Typography></div>
          <div className={this.props.classes.headerColumn}><Typography variant="body2">{devID}</Typography></div>
        </ExpansionPanelSummary>
        <ExpansionPanelDetails>
          <Grid container spacing={24}>
            <Grid item xs className={this.props.classes.treeStyle}>
              {this.props.frame.uplinkMetaData && <JSONTree data={this.props.frame.uplinkMetaData} />}
              {this.props.frame.downlinkMetaData && <JSONTree data={this.props.frame.downlinkMetaData} />}
            </Grid>
            <Grid item xs className={this.props.classes.treeStyle}>
              <JSONTree data={{phyPayload: this.props.frame.phyPayload}} />
            </Grid>
          </Grid>
        </ExpansionPanelDetails>
      </ExpansionPanel>
    );
  }
}

export default withStyles(styles)(LoRaWANFrameLog);
