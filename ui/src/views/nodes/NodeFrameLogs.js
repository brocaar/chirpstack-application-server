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

    return(
      <tr>
        <td>
          <span className={`glyphicon glyphicon-arrow-${dir}`} aria-hidden="true"></span>
        </td>
        <td>{createdAt}</td>
        <td>
          <JSONTree data={rxtx} theme={theme} hideRoot={true} />
        </td>
        <td>
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
    const page = (props.location.query.page === undefined) ? 1 : props.location.query.page;

    NodeStore.getFrameLogs(this.props.params.devEUI, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, frames) => {
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
          <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/nodes/${this.props.params.devEUI}/frames`} />
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
