import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import Chip from "@material-ui/core/Chip";
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
    marginRight: 6 * theme.spacing(1),
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
  chip: {
    marginRight: theme.spacing(1),
  },
};


class LoRaWANFrameLog extends Component {
  constructor() {
    super();

    this.state = {
      expanded: false,
    };

  }

  onExpandChange = (e, expanded) => {
    this.setState({
      expanded: expanded,
    });
  } 

  render() {
    const receivedAt = moment(this.props.frame.publishedAt).format("MMM DD LTS");

    let chips = [
      <Chip className={this.props.classes.chip} size="small" variant="outlined" label={this.props.frame.phyPayload.mhdr.mType} />,
      <Chip className={this.props.classes.chip} size="small" variant="outlined" label={`${this.props.frame.txInfo.frequency / 1000000} MHz`} />,
    ];

    if (this.props.frame.txInfo !== undefined && this.props.frame.txInfo.loRaModulationInfo !== undefined) {
      chips.push(<Chip className={this.props.classes.chip} size="small" variant="outlined" label={`SF${this.props.frame.txInfo.loRaModulationInfo.spreadingFactor}`} />);
      chips.push(<Chip className={this.props.classes.chip} size="small" variant="outlined" label={`BW${this.props.frame.txInfo.loRaModulationInfo.bandwidth}`} />);
    }

    if (this.props.frame.phyPayload.macPayload !== undefined) {
      if (this.props.frame.phyPayload.macPayload.devEUI !== undefined) {
        chips.push(<Chip className={this.props.classes.chip} size="small" variant="outlined" label={`DevEUI: ${this.props.frame.phyPayload.macPayload.devEUI}`} />);
      }

      if (this.props.frame.phyPayload.macPayload.fPort !== undefined && this.props.frame.phyPayload.macPayload.fPort !== null) {
        chips.push(<Chip className={this.props.classes.chip} size="small" variant="outlined" label={`FPort: ${this.props.frame.phyPayload.macPayload.fPort}`} />);
      }

      if (this.props.frame.phyPayload.macPayload.fhdr !== undefined) {
        chips.push(<Chip className={this.props.classes.chip} size="small" variant="outlined" label={`FCnt: ${this.props.frame.phyPayload.macPayload.fhdr.fCnt}`} />);
        chips.push(<Chip className={this.props.classes.chip} size="small" variant="outlined" label={`DevAddr: ${this.props.frame.phyPayload.macPayload.fhdr.devAddr}`} />);
      }
    }

    if (this.props.frame.gatewayID !== undefined) {
        chips.push(<Chip className={this.props.classes.chip} size="small" variant="outlined" label={`GW: ${this.props.frame.gatewayID}`} />);
    }

    return(
      <ExpansionPanel onChange={this.onExpandChange}>
        <ExpansionPanelSummary expandIcon={<ChevronDown />}>
          <div className={this.props.classes.headerColumnFixedSmall}><Typography variant="body2">{receivedAt}</Typography></div>
          {chips}
        </ExpansionPanelSummary>
        <ExpansionPanelDetails>
          {this.state.expanded && <Grid container spacing={4}>
            <Grid item xs className={this.props.classes.treeStyle}>
              {this.props.frame.txInfo && <JSONTree data={{gatewayID: this.props.frame.gatewayID, txInfo: this.props.frame.txInfo}} />}
              {this.props.frame.rxInfo && <JSONTree data={{rxInfo: this.props.frame.rxInfo}} />}
            </Grid>
            <Grid item xs className={this.props.classes.treeStyle}>
              <JSONTree data={{phyPayload: this.props.frame.phyPayload}} />
            </Grid>
          </Grid>}
        </ExpansionPanelDetails>
      </ExpansionPanel>
    );
  }
}

export default withStyles(styles)(LoRaWANFrameLog);
