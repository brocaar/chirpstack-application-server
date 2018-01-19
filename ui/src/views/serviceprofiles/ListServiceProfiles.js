import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import Pagination from "../../components/Pagination";
import ServiceProfileStore from "../../stores/ServiceProfileStore";
import SessionStore from "../../stores/SessionStore";


class ServiceProfileRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/organizations/${this.props.organizationID}/service-profiles/${this.props.serviceProfile.serviceProfileID}`}>{this.props.serviceProfile.name}</Link></td>
      </tr>
    );
  }
}

class ListServiceProfiles extends Component {
  constructor() {
    super();

    this.state = {
      pageSize: 20,
      serviceProfiles: [],
      isAdmin: false,
      pageNumber: 1,
      pages: 1,
    };

    this.updatePage = this.updatePage.bind(this);
  }

  componentDidMount() {
    this.updatePage(this.props);

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: SessionStore.isAdmin(),
      });
    });
  }

  updatePage(props) {
    this.setState({
      isAdmin: SessionStore.isAdmin(),
    });

    const query = new URLSearchParams(props.location.search);
    const page = (query.get('page') === null) ? 1 : query.get('page');

    ServiceProfileStore.getAllForOrganizationID(props.match.params.organizationID, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, serviceProfiles) => {
      this.setState({
        serviceProfiles: serviceProfiles,
        pageNumber: page,
        pages: Math.ceil(totalCount / this.state.pageSize),
      });
      window.scrollTo(0, 0);
    });
  }

  render() {
    const ServiceProfileRows = this.state.serviceProfiles.map((serviceProfile, i) => <ServiceProfileRow key={serviceProfile.serviceProfileID} serviceProfile={serviceProfile} organizationID={this.props.match.params.organizationID} />);

    return(
      <div className="panel panel-default">
        <div className={`panel-heading clearfix ${this.state.isAdmin ? '' : 'hidden'}`}>
          <div className="btn-group pull-right">
            <Link to={`/organizations/${this.props.match.params.organizationID}/service-profiles/create`}><button type="button" className="btn btn-default btn-sm">Create service-profile</button></Link>
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
              {ServiceProfileRows}
            </tbody>
          </table>
        </div>
        <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.match.params.organizationID}/service-profiles`} />
      </div>
    );
  }
}

export default ListServiceProfiles;