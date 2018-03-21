import React, { Component } from "react";
import JSONTree from "react-json-tree";
import moment from "moment";
import fileDownload from "js-file-download";

import GatewayStore from "../../stores/GatewayStore";


class FrameRow extends Component {
  render() {
    const theme = {
      scheme: 'google',
      author: 'seth wright (http://sethawright.com)',
      base00: '#1d1f21',
      base01: '#282a2e',
      base02: '#373b41',
      base03: '#969896',
      base04: '#b4b7b4',
      base05: '#c5c8c6',
      base06: '#e0e0e0',
      base07: '#ffffff',
      base08: '#CC342B',
      base09: '#F96A38',
      base0A: '#FBA922',
      base0B: '#198844',
      base0C: '#3971ED',
      base0D: '#3971ED',
      base0E: '#A36AC7',
      base0F: '#3971ED',
    }

    const data = {
      phyPayload: this.props.frame.phyPayload,
    };

    let rxtx = {};
    let dir = "";

    if (this.props.frame.uplinkMetaData !== undefined) {
      dir = "up";
      rxtx["uplink"] = this.props.frame.uplinkMetaData;
    }

    if (this.props.frame.downlinkMetaData !== undefined) {
      dir = "down";
      rxtx["downlink"] = this.props.frame.downlinkMetaData;
    }

    const receivedAt = moment(this.props.frame.receivedAt).format('LTS');
    const treeStyle = {
      paddingTop: '0',
      paddingBottom: '0',
    };

    return(
      <tr>
        <td>
          <span className={`glyphicon glyphicon-arrow-${dir}`} aria-hidden="true"></span>
        </td>
        <td>{receivedAt}</td>
        <td style={treeStyle}>
          <JSONTree data={rxtx} theme={theme} hideRoot={true} />
        </td>
        <td style={treeStyle}>
          <JSONTree data={data} theme={theme} hideRoot={true} />
        </td>
      </tr>
    );
  }
}

class GatewayFrameLogs extends Component {
  constructor() {
    super();
    this.state = {
      wsConnected: false,
      frames: [],
      paused: false,
    };

    this.onConnected = this.onConnected.bind(this);
    this.onDisconnected = this.onDisconnected.bind(this);
    this.onFrame = this.onFrame.bind(this);
    this.togglePause = this.togglePause.bind(this);
    this.clearFrames = this.clearFrames.bind(this);
    this.download = this.download.bind(this);
  }

  togglePause() {
    this.setState({
      paused: !this.state.paused,
    });
  }

  clearFrames() {
    this.setState({
      frames: [],
    });
  }

  download() {
    const dl = this.state.frames.map((frame, i) => {
      return {
        uplinkMetaData: frame.uplinkMetaData,
        downlinkMetaData: frame.downlinkMetaData,
        phyPayload: frame.phyPayload,
      }
    });

    fileDownload(JSON.stringify(dl, null, 4), `gateway-${this.props.match.params.mac}.json`);
  }

  onConnected() {
    this.setState({
      wsConnected: true,
    });
  }

  onDisconnected() {
    this.setState({
      wsConnected: false,
    });
  }

  onFrame(frame) {
    if (this.state.paused) {
      return;
    }

    let frames = this.state.frames;
    const now = new Date();

    if (frame.uplinkFrames.length !== 0) {
      frames.unshift({
        id: now.getTime(),
        receivedAt: new Date(),
        uplinkMetaData: {
          rxInfo: frame.uplinkFrames[0].rxInfo,
          txInfo: frame.uplinkFrames[0].txInfo,
        },
        phyPayload: JSON.parse(frame.uplinkFrames[0].phyPayloadJSON),
      });
    }

    if (frame.downlinkFrames.length !== 0) {
      frames.unshift({
        id: now.getTime(),
        receivedAt: new Date(),
        downlinkMetaData: {
          txInfo: frame.downlinkFrames[0].txInfo,
        },
        phyPayload: JSON.parse(frame.downlinkFrames[0].phyPayloadJSON),
      });
    }

    this.setState({
      frames: frames,
    });

    console.log(frame);
  }

  componentDidMount() {
    const conn = GatewayStore.getFrameLogsConnection(this.props.match.params.mac, this.onConnected, this.onDisconnected, this.onFrame);
    this.setState({
      wsConn: conn,
    });
  }

  componentWillUnmount() {
    this.state.wsConn.close();
  }

  render() {
    const FrameRows = this.state.frames.map((frame, i) => <FrameRow key={frame.id} frame={frame} />);
    let status;

    if (this.state.wsConnected) {
      status = <span className="label label-success">connected</span>;
    } else {
      status = <span className="label label-danger">disconnected</span>;
    }

    return (
      <div>
        <div className="panel panel-default">
          <div className="panel-heading clearfix">
            <h3 className="panel-title panel-title-buttons pull-left">Live LoRaWAN frame logs {status}</h3> 
            <div className="btn-group pull-right" role="group" aria-label="...">
              <button type="button" className={`btn btn-default btn-sm ${this.state.paused ? 'hidden': ''}`} onClick={this.togglePause}>
                <span className="glyphicon glyphicon-pause" aria-hidden="true"></span> Pause
              </button>
              <button type="button" className={`btn btn-default btn-sm ${this.state.paused ? '': 'hidden'}`} onClick={this.togglePause}>
                <span className="glyphicon glyphicon-play" aria-hidden="true"></span> Start
              </button>
              <button type="button" className="btn btn-default btn-sm" onClick={this.clearFrames}>
                <span className="glyphicon glyphicon-trash" aria-hidden="true"></span> Clear logs
              </button>
              <button type="button" className="btn btn-default btn-sm" onClick={this.download}>
                <span className="glyphicon glyphicon-download" aria-hidden="true"></span> Download
              </button>
            </div>
          </div>
          <div className="panel-body">
            <div className="alert alert-warning">
              The frames below are the raw (and encrypted) LoRaWAN PHYPayload frames as seen by the gateway.
              This data is inteded for debugging only.
            </div>
            <table className="table">
              <thead>
                <tr>
                  <th className="col-md-1">&nbsp;</th>
                  <th className="col-md-2">Received</th>
                  <th className="col-md-4">RX / TX parameters</th>
                  <th>LoRaWAN PHYPayload</th>
                </tr>
              </thead>
              <tbody>
                {FrameRows}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}

export default GatewayFrameLogs;
