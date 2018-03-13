import React, { Component } from 'react';
import moment from "moment";
import { Link } from 'react-router-dom';

import Pagination from "../../components/Pagination";
import NodeStore from "../../stores/NodeStore";
import SessionStore from "../../stores/SessionStore";
import ApplicationStore from "../../stores/ApplicationStore";


class NodeRow extends Component {
  render() {
    let lastseen = "n/a";
    let margin = "n/a";
    let battery = "n/a";
    if (this.props.node.lastSeenAt !== undefined && this.props.node.lastSeenAt !== "") {
      lastseen = moment(this.props.node.lastSeenAt).fromNow();
    }

    if (this.props.node.deviceStatusMargin !== undefined && this.props.node.deviceStatusMargin !== 256) {
      margin = `${this.props.node.deviceStatusMargin} dB`;
    }

    if (this.props.node.deviceStatusBattery !== undefined && this.props.node.deviceStatusBattery !== 256) {
      if (this.props.node.deviceStatusBattery === 255) {
        battery = "n/a";
      } else if (this.props.node.deviceStatusBattery === 0) {
        battery = "external";
      } else {
        battery = Math.round(100/255*this.props.node.deviceStatusBattery) + " %";
      }
    }

    return(
      <tr>
        <td>{lastseen}</td>
        <td><Link to={`/organizations/${this.props.application.organizationID}/applications/${this.props.application.id}/nodes/${this.props.node.devEUI}/edit`}>{this.props.node.name}</Link></td>
        <td>{this.props.node.devEUI}</td>
        <td><Link to={`/organizations/${this.props.application.organizationID}/device-profiles/${this.props.node.deviceProfileID}`}>{this.props.node.deviceProfileName}</Link></td>
        <td>{margin}</td>
        <td>{battery}</td>
      </tr>
    );
  }
}

class ListNodes extends Component {
  constructor() {
    super();
    this.state = {
      application: {},
      nodes: [],
      isAdmin: false,
      pageSize: 20,
      pageNumber: 1,
      pages: 1,
      search: "",
    };

    this.updatePage = this.updatePage.bind(this);
    this.onChange = this.onChange.bind(this);
    this.searchNodes = this.searchNodes.bind(this);
  }

  componentDidMount() {
    this.updatePage(this.props);
    ApplicationStore.getApplication(this.props.match.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
      });
    });
  }

  componentWillReceiveProps(nextProps) {
    this.updatePage(nextProps);
  }


  updatePage(props) {
    const query = new URLSearchParams(props.location.search);
    const page = (query.get('page') === null) ? 1 : query.get('page');

    NodeStore.getAll(this.props.match.params.applicationID, this.state.pageSize, (page-1) * this.state.pageSize, this.state.search, (totalCount, nodes) => {
      this.setState({
        nodes: nodes,
        pageNumber: page,
        pages: Math.ceil(totalCount / this.state.pageSize),
      });
      window.scrollTo(0, 0);
    });
  }

  onChange(e) {
    this.setState({
      search: e.target.value,
    });
    
  }

  searchNodes(e) {
    e.preventDefault();
    this.updatePage(this.props);
  }

  render() {
    const NodeRows = this.state.nodes.map((node, i) => <NodeRow key={node.devEUI} node={node} application={this.state.application} />);

    const searchStyle = {
      width: "200px",
    };

    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <form className="form-inline pull-left" onSubmit={this.searchNodes}>
            <div className="form-group">
              <div className="input-group">
                <div className="input-group-addon">
                  <span className="glyphicon glyphicon-search" aria-hidden="true"></span>
                </div>
                <input type="text" className="form-control" style={searchStyle} placeholder="Device name or DevEUI" onChange={this.onChange} value={this.state.search || ''} />
              </div>
            </div>
          </form>
          <div className={`btn-group pull-right ${this.state.isAdmin ? "" : "hidden"}`}>
            <Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.state.application.id}/nodes/create`}><button type="button" className="btn btn-default btn-sm">Create device</button></Link>
          </div>
        </div>
        <div className="panel-body">
          <table className="table table-hover">
            <thead>
              <tr>
                <th className="col-md-2">Last seen</th>
                <th className="col-md-3">Device name</th>
                <th className="col-md-2">Device EUI</th>
                <th className="col-md-3">Device-profile</th>
                <th className="col-md-1">Link margin</th>
                <th className="col-md-1">Battery</th>
              </tr>
            </thead>
            <tbody>
              {NodeRows}
            </tbody>
          </table>
        </div>
        <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`} />
      </div>
    );
  }
}

export default ListNodes;
