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
        <td><Link to={`/applications/${this.props.application.id}/nodes/${this.props.node.devEUI}/edit`}>{this.props.node.name}</Link></td>
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
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      application: {},
      nodes: [],
      isApplicationAdmin: false,
      pageSize: 20,
      pageNumber: 1,
      pages: 1,
    };

    this.onDelete = this.onDelete.bind(this);
    this.updatePage = this.updatePage.bind(this);
  }

  componentWillMount() {
    this.updatePage(this.props);
    ApplicationStore.getApplication(this.props.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isApplicationAdmin: (SessionStore.isAdmin() || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isApplicationAdmin: (SessionStore.isAdmin() || SessionStore.isApplicationAdmin(this.props.params.applicationID)),
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

  onDelete() {
    if (confirm("Are you sure you want to delete this application?")) {
      ApplicationStore.deleteApplication(this.props.params.applicationID, (responseData) => {
        this.context.router.push("/applications");
      }); 
    }
  }

  render() {
    const NodeRows = this.state.nodes.map((node, i) => <NodeRow key={node.devEUI} node={node} application={this.state.application} />);

    return(
      <div>
        <ol className="breadcrumb">
          <li><Link to="/">Dashboard</Link></li>
          <li><Link to="/applications">Applications</Link></li>
          <li>{this.state.application.name}</li>
          <li className="active">Nodes</li>
        </ol>
        <div className={(this.state.isApplicationAdmin ? '' : 'hidden')}>
          <div className="clearfix">
            <div className="btn-group pull-right" role="group" aria-label="...">
              <Link to={`/applications/${this.props.params.applicationID}/users`}><button type="button" className="btn btn-default">Users</button></Link> &nbsp;
              <Link to={`/applications/${this.props.params.applicationID}/nodes/create`}><button type="button" className="btn btn-default">Create node</button></Link> &nbsp;
              <Link to={`/applications/${this.props.params.applicationID}/edit`}><button type="button" className="btn btn-default">Edit application</button></Link> &nbsp;
              <Link><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete application</button></Link>
            </div>
          </div>
        </div>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
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
          <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/applications/${this.props.params.applicationID}`} />
        </div>
      </div>
    );
  }
}

export default ListNodes;
