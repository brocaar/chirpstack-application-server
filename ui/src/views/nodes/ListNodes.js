import React, { Component } from 'react';
import { Link } from 'react-router';

import Pagination from "../../components/Pagination";
import NodeStore from "../../stores/NodeStore";
import SessionStore from "../../stores/SessionStore";
import ApplicationStore from "../../stores/ApplicationStore";

class NodeRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/organizations/${this.props.application.organizationID}/applications/${this.props.application.id}/nodes/${this.props.node.devEUI}/edit`}>{this.props.node.name}</Link></td>
        <td>{this.props.node.devEUI}</td>
        <td>{this.props.node.description}</td>
        <td>
          <span className={this.props.node.isABP ? 'hidden' : ''}>OTAA</span>
          <span className={this.props.node.isABP ? '' : 'hidden'}>ABP</span>
        </td>
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
  }

  componentDidMount() {
    this.updatePage(this.props);
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.params.organizationID) || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
      });
    });
  }

  componentWillReceiveProps(nextProps) {
    this.updatePage(nextProps);
  }


  updatePage(props) {
    const page = (props.location.query.page === undefined) ? 1 : props.location.query.page;

    NodeStore.getAll(this.props.params.applicationID, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, nodes) => {
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
    this.updatePage(this.props);
  }

  render() {
    const NodeRows = this.state.nodes.map((node, i) => <NodeRow key={node.devEUI} node={node} application={this.state.application} />);

    return(
      <div className="panel panel-default">
        <div className={`panel-heading clearfix ${this.state.isAdmin ? '' : 'hidden'}`}>
          <div className="btn-group pull-right">
            <Link to={`/organizations/${this.props.params.organizationID}/applications/${this.state.application.id}/nodes/create`}><button type="button" className="btn btn-default btn-sm">Create node</button></Link>
          </div>
        </div>
        <div className="panel-body">
          <div className="input-group col-md-3">
            <input type="text" className="form-control" placeholder="Search for..." onChange={this.onChange} />
          </div>
          <table className="table table-hover">
            <thead>
              <tr>
                <th className="col-md-3">Device name</th>
                <th className="col-md-2">Device EUI</th>
                <th>Device description</th>
                <th className="col-md-1">Activation</th>
              </tr>
            </thead>
            <tbody>
              {NodeRows}
            </tbody>
          </table>
        </div>
        <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}`} />
      </div>
    );
  }
}

export default ListNodes;
