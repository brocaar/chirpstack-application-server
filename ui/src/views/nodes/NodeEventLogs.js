import React, { Component } from "react";
import JSONTree from "react-json-tree";
import moment from "moment";
import fileDownload from "js-file-download";

import NodeStore from "../../stores/NodeStore";


class EventRow extends Component {
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

    const receivedAt = moment(this.props.event.receivedAt).format("LTS");
    const treeStyle = {
      paddingTop: '0',
      paddingBottom: '0',
    };

    const data = {
      payload: this.props.event.payload,
    };

    return(
      <tr>
        <td>{receivedAt}</td>
        <td>{this.props.event.type}</td>
        <td style={treeStyle}>
          <JSONTree data={data} theme={theme} hideRoot={true} />
        </td>
      </tr>
    );
  }
}


class NodeEventLogs extends Component {
  constructor() {
    super();

    this.state = {
      wsConnected: false,
      events: [],
      paused: false,
    };

    this.onConnected = this.onConnected.bind(this);
    this.onDisconnected = this.onDisconnected.bind(this);
    this.onEvent = this.onEvent.bind(this);
    this.togglePause = this.togglePause.bind(this);
    this.clearEvents = this.clearEvents.bind(this);
    this.download = this.download.bind(this);
  }

  togglePause() {
    this.setState({
      paused: !this.state.paused,
    });
  }

  clearEvents() {
    this.setState({
      events: [],
    });
  }

  download() {
    const dl = this.state.events.map((event, i) => {
      return {
        type: event.type,
        payload: event.payload,
      }
    });

    fileDownload(JSON.stringify(dl, null, 4), `device-${this.props.match.params.devEUI}.json`);
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

  onEvent(event) {
    if (this.state.paused) {
      return;
    }

    let events = this.state.events;
    const now = new Date();

    events.unshift({
      id: now.getTime(),
      receivedAt: new Date(),
      type: event.type,
      payload: JSON.parse(event.payloadJSON),
    });

    this.setState({
      events: events,
    });
  }

  componentDidMount() {
    const conn = NodeStore.getEventLogsConnection(this.props.match.params.devEUI, this.onConnected, this.onDisconnected, this.onEvent);
    this.setState({
      wsConn: conn,
    });
  }

  componentWillUnmount() {
    this.state.wsConn.close();
  }

  render() {
    const eventRows = this.state.events.map((event, i) => <EventRow key={event.id} event={event} />);
    let status;

    if (this.state.wsConnected) {
      status = <span className="label label-success">connected</span>;
    } else {
      status = <span className="label label-danger">disconnected</span>;
    }

    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-heading clearfix">
            <h3 className="panel-title panel-title-buttons pull-left">Live event logs {status}</h3> 
            <div className="btn-group pull-right" role="group" aria-label="...">
              <button type="button" className={`btn btn-default btn-sm ${this.state.paused ? 'hidden': ''}`} onClick={this.togglePause}>
                <span className="glyphicon glyphicon-pause" aria-hidden="true"></span> Pause
              </button>
              <button type="button" className={`btn btn-default btn-sm ${this.state.paused ? '': 'hidden'}`} onClick={this.togglePause}>
                <span className="glyphicon glyphicon-play" aria-hidden="true"></span> Start
              </button>
              <button type="button" className="btn btn-default btn-sm" onClick={this.clearEvents}>
                <span className="glyphicon glyphicon-trash" aria-hidden="true"></span> Clear logs
              </button>
              <button type="button" className="btn btn-default btn-sm" onClick={this.download}>
                <span className="glyphicon glyphicon-download" aria-hidden="true"></span> Download
              </button>
            </div>
          </div>
          <div className="panel-body">
            <div className="alert alert-warning">
              These are the events as published to the application. Please refer to <a href="https://www.loraserver.io/lora-app-server/integrate/integrations/">data integrations</a> for more information on integrating this with your application.
            </div>
            <table className="table">
              <thead>
                <tr>
                  <th className="col-md-2">Received</th>
                  <th className="col-md-2">Event type</th>
                  <th>Event payload</th>
                </tr>
              </thead>
              <tbody>
                {eventRows}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}

export default NodeEventLogs;
