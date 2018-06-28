import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import Pagination from "../../components/Pagination";
import GatewayProfileStore from "../../stores/GatewayProfileStore";


class GatewayProfileRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/network-servers/${this.props.profile.networkServerID}/gateway-profiles/${this.props.profile.id}/edit`}>{this.props.profile.name}</Link></td>
      </tr>
    );
  }
}

class ListGatewayProfiles extends Component {
  constructor() {
    super();

    this.state = {
      pageSize: 20,
      gatewayProfiles: [],
      pageNumber: 1,
      pages: 1,
    };

    this.updatePage = this.updatePage.bind(this);
  }

  componentDidMount() {
    this.updatePage(this.props);
  }

  updatePage(props) {
    const query = new URLSearchParams(props.location.search);
    const page = (query.get('page') === null) ? 1 : query.get('page');

    GatewayProfileStore.getAllForNetworkServerID(this.props.match.params.networkServerID, this.state.pageSize, (page-1) * this.state.pageSize, (count, profiles) => {
      this.setState({
        gatewayProfiles: profiles,
        pageNumber: page,
        pages: Math.ceil(count/this.state.pageSize),
      });
      window.scrollTo(0, 0);
    });
  }

  render() {
    const profileRows = this.state.gatewayProfiles.map((gp, i) => <GatewayProfileRow key={gp.gatewayProfileID} profile={gp} />);

    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-heading clearfix">
            <div className="btn-group pull-right">
              <Link to={`/network-servers/${this.props.match.params.networkServerID}/gateway-profiles/create`}><button type="button" className="btn btn-default btn-sm">Create gateway-profile</button></Link>
            </div>
          </div>
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>Name</th>
                </tr>
              </thead>
              <tbody>
                {profileRows}
              </tbody>
            </table>
          </div>
        </div>
        <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/network-servers/${this.props.match.params.networkServerID}/gateway-profiles`} />
      </div>
    );
  }
}

export default ListGatewayProfiles;
