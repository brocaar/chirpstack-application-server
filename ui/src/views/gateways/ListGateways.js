import React, { Component } from 'react';
import { Link  } from 'react-router-dom';

import moment from "moment";
import { Bar } from "react-chartjs";

import Pagination from "../../components/Pagination";
import GatewayStore from "../../stores/GatewayStore";
import SessionStore from "../../stores/SessionStore";


class GatewayRow extends Component {
  constructor() {
    super();

    this.state = {
      stats: {
        labels: [],
          datasets: [
            {
              data: [],
              fillColor: "rgba(33, 150, 243, 1)",
            },
          ],
      },
      options: {
        animation: false,
        showScale: false,
        showTooltips: false,
        barShowStroke: false,
        barValueSpacing: 5,
      },
    };
  }


  componentWillMount() {
    GatewayStore.getGatewayStats(this.props.gateway.mac, "DAY", moment().subtract(29, 'days').toISOString(), moment().toISOString(), (records) => {
      let stats = this.state.stats;
      for (const record of records) {
        stats.labels.push(record.timestamp);
        stats.datasets[0].data.push(record.rxPacketsReceived + record.txPacketsReceived);
      }

      this.setState({
        stats: stats,
      });
    });
  }

  render() {
    return(
      <tr>
        <td><Link to={`/organizations/${this.props.gateway.organizationID}/gateways/${this.props.gateway.mac}`}>{this.props.gateway.name}</Link></td>
        <td>{this.props.gateway.mac}</td>
        <td>
          <Bar width="380" height="23" data={this.state.stats} options={this.state.options} />
        </td>
      </tr>
    );
  }
}

class ListGateways extends Component {
  constructor() {
    super();

    this.state = {
      gateways: [],
      pageSize: 20,
      pageNumber: 1,
      pages: 1,
      isAdmin: false,
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

  componentWillReceiveProps(nextProps) {
    this.updatePage(nextProps);
  }

  updatePage(props) {
    this.setState({
      isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(props.match.params.organizationID),
    });

    const query = new URLSearchParams(props.location.search);
    const page = (query.get('page') === null) ? 1 : query.get('page');

    GatewayStore.getAllForOrganization(props.match.params.organizationID, this.state.pageSize, (page-1) * this.state.pageSize, (totalCount, gateways) => {
      this.setState({
        gateways: gateways,
        pageNumber: page,
        pages: Math.ceil(totalCount / this.state.pageSize),
      });
    });

    window.scrollTo(0, 0);
  }

  render() {
    const GatewayRows = this.state.gateways.map((gw, i) => <GatewayRow key={gw.mac} gateway={gw} />);

    return(
      <div className="panel panel-default">
        <div className={`panel-heading clearfix ${this.state.isAdmin ? '' : 'hidden'}`}>
          <div className="btn-group pull-right">
            <Link to={`/organizations/${this.props.match.params.organizationID}/gateways/create`}><button type="button" className="btn btn-default btn-sm">Create gateway</button></Link>
          </div>
        </div>
        <div className="panel-body">
          <table className="table table-hover">
            <thead>
              <tr>
                <th className="col-md-3">Name</th>
                <th>MAC</th>
                <th className="col-md-4">Gateway activity (30d)</th>
              </tr>
            </thead>
            <tbody>
              {GatewayRows}
            </tbody>
          </table>
        </div>
        <Pagination pages={this.state.pages} currentPage={this.state.pageNumber} pathname={`/organizations/${this.props.match.params.organizationID}/gateways`} />
      </div>
    );
  }
}

export default ListGateways;
