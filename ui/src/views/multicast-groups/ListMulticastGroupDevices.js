import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Button from '@material-ui/core/Button';
import Checkbox from "@material-ui/core/Checkbox";
import Delete from "mdi-material-ui/Delete";

import DeviceAdmin from "../../components/DeviceAdmin";
import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";
import DeviceStore from "../../stores/DeviceStore";
import theme from "../../theme";
import MulticastGroupStore from "../../stores/MulticastGroupStore";


const styles = {
  select: {
    width: 2 * theme.spacing(1),
    padding: 0,
  },
  buttons: {
    textAlign: "right",
  },
  button: {
    marginLeft: 2 * theme.spacing(1),
  },
  icon: {
    marginRight: theme.spacing(1),
  },
};



class ListMulticastGroupDevices extends Component {
  constructor() {
    super();

    this.state = {
      selected: {},
    };
  }

  getPage = (limit, offset, callbackFunc) => {
    DeviceStore.list({
      multicastGroupID: this.props.match.params.multicastGroupID,
      limit: limit,
      offset: offset,
    }, callbackFunc);
  }

  onCheckboxChange = (e) => {
    let selected = this.state.selected;

    if (e.target.checked) {
      selected[e.target.id] = true;
    } else {
      delete selected[e.target.id];
    }
    
    this.setState({
      selected: selected,
    });
  }

  removeDevices = () => {
    if (window.confirm("Are you sure you want to remove the selected devices from the multicast-group?")) {
      let count = 0;
      let self = this;

      for (const [key, ] of Object.entries(this.state.selected)) {
        count++;

        MulticastGroupStore.removeDevice(this.props.match.params.multicastGroupID, key, function (cnt) {
          return function () {
            // reload after the last request completed
            if (cnt === Object.keys(self.state.selected).length) {
              self.setState({
                selected: {},
              });
              self.forceUpdate();
            }
          };
        }(count));
      }
    }
  }

  getRow = (obj) => {
    return(
      <TableRow
        key={obj.devEUI}
        hover
      >
        <TableCell className={this.props.classes.select}>
          <DeviceAdmin organizationID={this.props.match.params.organizationID}>
            <Checkbox
              id={obj.devEUI}
              onChange={this.onCheckboxChange}
              checked={!!this.state.selected[obj.devEUI]}
            />
          </DeviceAdmin>
        </TableCell>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${obj.applicationID}/devices/${obj.devEUI}`}>{obj.name}</TableCellLink>
        <TableCell>{obj.devEUI}</TableCell>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={4}>
        <DeviceAdmin organizationID={this.props.match.params.organizationID}>
          <Grid item xs={12} className={this.props.classes.buttons}>
            <Button variant="outlined" disabled={Object.keys(this.state.selected).length === 0} color="secondary" className={this.props.classes.button} onClick={this.removeDevices}>
              <Delete className={this.props.classes.icon} />
              Remove from group
            </Button>
          </Grid>
        </DeviceAdmin>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell></TableCell>
                <TableCell>Device name</TableCell>
                <TableCell>Device EUI</TableCell>
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

export default withStyles(styles)(ListMulticastGroupDevices);
