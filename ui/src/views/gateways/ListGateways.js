import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

import Plus from "mdi-material-ui/Plus";

import moment from "moment";
import { Bar } from "react-chartjs-2";

import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import TableCellLink from "../../components/TableCellLink";
import TitleBarButton from "../../components/TitleBarButton";
import DataTable from "../../components/DataTable";
import Admin from "../../components/Admin";
import GatewayStore from "../../stores/GatewayStore";


class GatewayRow extends Component {
  constructor() {
    super();

    this.state = {};
  }

  componentWillMount() {
    const start = moment().subtract(29, "days").toISOString();
    const end = moment().toISOString();

    GatewayStore.getStats(this.props.gateway.id, start, end, resp => {
      let stats = {
        labels: [],
        datasets: [
          {
            data: [],
            fillColor: "rgba(33, 150, 243, 1)",
          },
        ],
      };

      for (const row of resp.result) {
        stats.labels.push(row.timestamp);
        stats.datasets[0].data.push(row.rxPacketsReceivedOK + row.txPacketsEmitted);
      }

      this.setState({
        stats: stats,
      });
    });
  }

  render() {
    const options = {
      scales: {
        xAxes: [{display: false}],
        yAxes: [{display: false}],
      },
      tooltips: {
        enabled: false,
      },
      legend: {
        display: false,
      },
      responsive: false,
      animation: {
        duration: 0,
      },
    };

    return(
      <TableRow>
          <TableCellLink to={`/organizations/${this.props.gateway.organizationID}/gateways/${this.props.gateway.id}`}>{this.props.gateway.name}</TableCellLink>
          <TableCell>{this.props.gateway.id}</TableCell>
          <TableCell>
            {this.state.stats && <Bar
              width={380}
              height={23}
              data={this.state.stats}
              options={options}
            />}
          </TableCell>
      </TableRow>
    );
  }
}


const styles = {
  chart: {
    width: 380,
  },
};



class ListGateways extends Component {
  constructor() {
    super();
    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  getPage(limit, offset, callbackFunc) {
    GatewayStore.list("", this.props.match.params.organizationID, limit, offset, callbackFunc);
  }

  getRow(obj) {
    return(
      <GatewayRow key={obj.id} gateway={obj} />
    );
  }

  render() {
    return(
      <Grid container spacing={24}>
        <TitleBar
          buttons={
            <Admin organizationID={this.props.match.params.organizationID}>
              <TitleBarButton
                key={1}
                label="Create"
                icon={<Plus />}
                to={`/organizations/${this.props.match.params.organizationID}/gateways/create`}
              />
            </Admin>
          }
        >
          <TitleBarTitle title="Gateways" />
        </TitleBar>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Gateway ID</TableCell>
                <TableCell className={this.props.classes.chart}>Gateway activity (30d)</TableCell>
              </TableRow>
            }
            getPage={this.getPage}
            getRow={this.getRow}
          />
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(ListGateways);
