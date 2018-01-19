import React, { Component } from "react";
import JSONTree from "react-json-tree";
import moment from "moment";

import NodeStore from "../../stores/NodeStore";
import Pagination from "../../components/Pagination";


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

    if (typeof(this.props.frame.txInfo) !== "undefined" && this.props.frame.txInfo !== null) {
      rxtx["txInfo"] = this.props.frame.txInfo;
      dir = "down";
    } else {
      rxtx["rxInfoSet"] = this.props.frame.rxInfoSet;
      dir = "up";
    }

    const createdAt = moment(this.props.frame.createdAt).format("LLLL");
    const treeStyle = {
      paddingTop: '0',
      paddingBottom: '0',
    };

    return(
      <tr>
        <td>
          <span className={`glyphicon glyphicon-arrow-${dir}`} aria-hidden="true"></span>
        </td>
        <td>{createdAt}</td>
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

class NodeFrameLogs extends Component {
  constructor() {
    super();

    this.state = {
      frames: [],
      pageSize: 20,
      pageNumber: 1,
      pages: 1,
    };

    this.updatePage = this.updatePage.bind(this);
  }

  componentDidMount() {
    this.updatePage(this.props);
  }

  componentWillReceiveProps(nextProps) {
    this.updatePage(nextProps);
  }

  updatePage(props) {
    const query = new URLSearchParams(props.location.search);
    const page = (query.get('page') === null) ? 1 : query.get('page');

    NodeStore.getFrameLogs(this.props.match.params.devEUI, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, frames) => {
      this.setState({
        frames: frames,
        pageNumber: page,
        pages: Math.ceil(totalCount / this.state.pageSize),
      });
      window.scrollTo(0, 0);
    });
  }

  render () {
    const FrameRows = this.state.frames.map((frame, i) => <FrameRow key={new Date().getTime() + i} frame={frame} />);

    if (FrameRows.length > 0) {
      return (
        <div>
          <div className="alert alert-warning" role="alert">
            The table below displays the raw and encrypted LoRaWAN frames. Use this data for debugging purposes.
            For application integration, please see the <a href="https://docs.loraserver.io/lora-app-server/integrate/data/">Send / receive data</a> documentation page.
          </div>
          <div className="panel panel-default">
            <div className="panel-body">
              <table className="table">
                <thead>
                  <tr>
                    <th className="col-md-1">&nbsp;</th>
                    <th className="col-md-3">Created at</th>
                    <th className="col-md-3">RX / TX parameters</th>
                    <th>Frame</th>
                  </tr>
                </thead>
                <tbody>
                  {FrameRows}
                </tbody>
              </table>
            </div>
            <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/nodes/${this.props.match.params.devEUI}/frames`} />
          </div>
        </div>
      );
    } else {
      return (
        <div className="panel panel-default">
          <div className="panel-body">
            No frames sent / received yet or LoRa Server has frame logging disabled.
          </div>
        </div>
      );
    }
  }
}

export default NodeFrameLogs;
