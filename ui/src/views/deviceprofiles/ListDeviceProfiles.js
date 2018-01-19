import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import Pagination from "../../components/Pagination";
import DeviceProfileStore from "../../stores/DeviceProfileStore";
import SessionStore from "../../stores/SessionStore";


class DeviceProfileRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/organizations/${this.props.organizationID}/device-profiles/${this.props.deviceProfile.deviceProfileID}`}>{this.props.deviceProfile.name}</Link></td>
      </tr>
    );
  }
}


class ListDeviceProfiles extends Component {
  constructor() {
    super();

    this.state = {
      pageSize: 20,
      deviceProfiles: [],
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
        isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID),
      });
    });
  }

  updatePage(props) {
    this.setState({
      isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID),
    });

    const query = new URLSearchParams(props.location.search);
    const page = (query.get('page') === null) ? 1 : query.get('page');

    DeviceProfileStore.getAllForOrganizationID(props.match.params.organizationID, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, deviceProfiles) => {
      this.setState({
        deviceProfiles: deviceProfiles,
        pageNumber: page,
        pages: Math.ceil(totalCount / this.state.pageSize),
      });
      window.scrollTo(0, 0);
    });
  }

  render() {
    const DeviceProfileRows = this.state.deviceProfiles.map((deviceProfile, i) => <DeviceProfileRow key={deviceProfile.deviceProfileID} deviceProfile={deviceProfile} organizationID={this.props.match.params.organizationID} />);

    return(
      <div className="panel panel-default">
        <div className={`panel-heading clearfix ${this.state.isAdmin ? '' : 'hidden'}`}>
          <div className="btn-group pull-right">
            <Link to={`/organizations/${this.props.match.params.organizationID}/device-profiles/create`}><button type="button" className="btn btn-default btn-sm">Create device-profile</button></Link>
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
              {DeviceProfileRows}
            </tbody>
          </table>
        </div>
        <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.match.params.organizationID}/device-profiles`} />
      </div>
    );
  }
}

export default ListDeviceProfiles;