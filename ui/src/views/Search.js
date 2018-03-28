import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import InternalStore from "../stores/InternalStore";


class ApplicationResult extends Component {
  render() {
    return(
      <tr>
        <td>application</td>
        <td><Link to={`/organizations/${this.props.result.organizationID}/applications/${this.props.result.applicationID}`}>{this.props.result.applicationName}</Link></td>
        <td>organization: <Link to={`/organizations/${this.props.result.organizationID}`}>{this.props.result.organizationName}</Link></td>
        <td>{this.props.result.applicationID}</td>
      </tr>
    );
  }
}

class OrganizationResult extends Component {
  render() {
    return(
      <tr>
        <td>organization</td>
        <td><Link to={`/organizations/${this.props.result.organizationID}`}>{this.props.result.organizationName}</Link></td>
        <td></td>
        <td>{this.props.result.organizationID}</td>
      </tr>
    );
  }
}

class DeviceResult extends Component {
  render() {
    return(
      <tr>
        <td>device</td>
        <td><Link to={`/organizations/${this.props.result.organizationID}/applications/${this.props.result.applicationID}/nodes/${this.props.result.deviceDevEUI}/edit`}>{this.props.result.deviceName}</Link></td>
        <td>organization: <Link to={`/organizations/${this.props.result.organizationID}`}>{this.props.result.organizationName}</Link>, application: <Link to={`/organizations/${this.props.result.organizationID}/applications/${this.props.result.applicationID}`}>{this.props.result.applicationName}</Link></td>
        <td>{this.props.result.deviceDevEUI}</td>
      </tr>
    );
  }
}

class GatewayResult extends Component {
  render() {
    return(
      <tr>
        <td>gateway</td>
        <td><Link to={`/organizations/${this.props.result.organizationID}/gateways/${this.props.result.gatewayMAC}`}>{this.props.result.gatewayName}</Link></td>
        <td>organization: <Link to={`/organizations/${this.props.result.organizationID}`}>{this.props.result.organizationName}</Link></td>
        <td>{this.props.result.gatewayMAC}</td>
      </tr>
    );
  }
}


class Search extends Component {
  constructor() {
    super();

    this.state = {
      results: [],
      pageSize: 20,
      pageNumber: 1,
    };

    this.updateSearch = this.updateSearch.bind(this);
  }

  componentDidMount() {
    this.updateSearch(this.props);
  }

  componentWillReceiveProps(nextProps) {
    this.updateSearch(nextProps);
  }

  updateSearch(props) {
    const query = new URLSearchParams(props.location.search);
    const search = (query.get('search') === null) ? "" : query.get('search');

    if (search === '') {
      return;
    }

    InternalStore.globalSearch(search, this.state.pageSize, 0, (results) => {
      this.setState({
        results: results,
      });
    });
  }

  render() {
    var searchResults = [];

    for (const result of this.state.results) {
      switch (result.kind) {
        case "application":
          searchResults.push(<ApplicationResult result={result} />);
          break;
        case "organization":
          searchResults.push(<OrganizationResult result={result} />);
          break;
        case "device":
          searchResults.push(<DeviceResult result={result} />);
          break;
        case "gateway":
          searchResults.push(<GatewayResult result={result} />);
          break;
        default:
          break;
      }
    }

    return(
      <div>        
        <ol className="breadcrumb">
          <li className="active">Search results</li>
        </ol>
        <hr />
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th className="col-md-2">Kind</th>
                  <th>Name</th>
                  <th></th>
                  <th className="col-md-3">ID</th>
                </tr>
              </thead>
              <tbody>
                {searchResults}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    );
  }
}

export default Search;
