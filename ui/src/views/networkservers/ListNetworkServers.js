import React, { Component } from "react";
import { Link } from 'react-router-dom';

import Pagination from "../../components/Pagination";
import NetworkServerStore from "../../stores/NetworkServerStore";


class NetworkServerRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/network-servers/${this.props.networkServer.id}`}>{this.props.networkServer.name}</Link></td>
        <td>{this.props.networkServer.server}</td>
      </tr>
    );
  }
}

class ListNetworkServers extends Component {
    constructor() {
      super();

      this.state = {
        pageSize: 20,
        networkServers: [],
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

      NetworkServerStore.getAll(this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, networkServers) => {
        this.setState({
          networkServers: networkServers,
          pageNumber: page,
          pages: Math.ceil(totalCount / this.state.pageSize),
        });
        window.scrollTo(0, 0);
      });
    }

    render() {
      const NetworkServerRows = this.state.networkServers.map((networkServer, i) => <NetworkServerRow key={networkServer.id} networkServer={networkServer} />);

      return(
        <div>
          <ol className="breadcrumb">
            <li className="active">Network servers</li>
          </ol>
          <hr />
          <div className="panel panel-default">
            <div className="panel-heading clearfix">
              <div className="btn-group pull-right">
                <Link to="/network-servers/create"><button type="button" className="btn btn-default btn-sm">Add network-server</button></Link>
              </div>
            </div>
            <div className="panel-body">
              <table className="table table-hover">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Server</th>
                  </tr>
                </thead>
                <tbody>
                  {NetworkServerRows}
                </tbody>
              </table>
            </div>
            <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname="/network-servers" />
          </div>
        </div>
      );
    }
}

export default ListNetworkServers;