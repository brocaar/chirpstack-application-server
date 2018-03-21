import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import Pagination from "../../components/Pagination";
import OrganizationStore from "../../stores/OrganizationStore";
import SessionStore from "../../stores/SessionStore";


class OrganizationRow extends Component {
  render() {
    return(
      <tr>
        <td>{this.props.organization.id}</td>
        <td><Link to={`/organizations/${this.props.organization.id}`}>{this.props.organization.name}</Link></td>
        <td>{this.props.organization.displayName}</td>
        <td><span className={"glyphicon glyphicon-" + (this.props.organization.canHaveGateways ? 'ok' : 'remove')} aria-hidden="true"></span></td>
      </tr>
    );
  }
}

class ListOrganizations extends Component {
  constructor() {
    super();

    this.state = {
      pageSize: 20,
      organizations: [],
      isAdmin: false,
      pageNumber: 1,
      pages: 1,
      search: "",
    };

    this.updatePage = this.updatePage.bind(this);
    this.onChange = this.onChange.bind(this);
    this.searchOrganizations = this.searchOrganizations.bind(this);
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

    OrganizationStore.getAll(this.state.search, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, organizations) => {
      this.setState({
    	organizations: organizations,
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

  searchOrganizations(e) {
    e.preventDefault();
    this.updatePage(this.props);
  }

  componentWillMount() {
    this.setState({
      isAdmin: SessionStore.isAdmin(),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: SessionStore.isAdmin(), 
      });
    });
  }

  render () {
    const OrganizationRows = this.state.organizations.map((organization, i) => <OrganizationRow key={organization.id} organization={organization} />);

    const searchStyle = {
      width: "200px",
    };

    return(
      <div>
        <ol className="breadcrumb">
          <li className="active">Organizations</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-heading clearfix">
            <form className="form-inline pull-left" onSubmit={this.searchOrganizations}>
              <div className="form-group">
                <div className="input-group">
                  <div className="input-group-addon">
                    <span className="glyphicon glyphicon-search" aria-hidden="true"></span>
                  </div>
                  <input type="text" className="form-control" style={searchStyle} placeholder="Organization display name" onChange={this.onChange} value={this.state.search || ''} />
                </div>
              </div>
            </form>
            <div className={`btn-group pull-right ${this.state.isAdmin ? "" : "hidden"}`}>
              <Link to="/organizations/create"><button type="button" className="btn btn-default btn-sm">Create organization</button></Link>
            </div>
          </div>
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th className="col-md-1">ID</th>
                  <th className="col-md-4">Name</th>
                  <th>Display name</th>
                  <th className="col-md-2">Can have gateways</th>
                </tr>
              </thead>
              <tbody>
                {OrganizationRows}
              </tbody>
            </table>
          </div>
          <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname="/organizations" />
        </div>
      </div>
    );
  }
}

export default ListOrganizations;
