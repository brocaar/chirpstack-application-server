import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Button from '@material-ui/core/Button';
import Checkbox from "@material-ui/core/Checkbox";
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import DialogTitle from '@material-ui/core/DialogTitle';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import Select from '@material-ui/core/Select';

import moment from "moment";
import Plus from "mdi-material-ui/Plus";
import PowerPlug from "mdi-material-ui/PowerPlug";

import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";
import DeviceAdmin from "../../components/DeviceAdmin";
import DeviceStore from "../../stores/DeviceStore";
import MulticastGroupStore from "../../stores/MulticastGroupStore";
import theme from "../../theme";


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


class ListDevices extends Component {
  constructor() {
    super();

    this.state = {
      selected: {},
      selectedMenuAnchor: null,
      multicastGroups: [],
      selectedMulticastGroup: "",
      multicastDialogOpen: false,
    };

    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  componentDidMount() {
    MulticastGroupStore.list("", this.props.match.params.applicationID, "", "", 999, 0, resp => {
      this.setState({
        multicastGroups: resp.result,
      });
    });
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

  getPage(limit, offset, callbackFunc) {
    DeviceStore.list({
      applicationID: this.props.match.params.applicationID,
      limit: limit,
      offset: offset,
    }, callbackFunc);
  }

  onSelectedMenuOpen = (e) => {
    this.setState({
      selectedMenuAnchor: e.currentTarget,
    });
  }

  onSelectedMenuClose = () => {
    this.setState({
      selectedMenuAnchor: null,
    });
  }

  onMulticastDialogClose = () => {
    this.setState({
      multicastDialogOpen: false,
    });
  }

  openMulticastDialog = () => {
    this.onSelectedMenuClose();

    this.setState({
      multicastDialogOpen: true,
      selectedMulticastGroup: "",
    });
  }
  
  onMulticastGroupSelectChange = (e) => {
    this.setState({
      selectedMulticastGroup: e.target.value,
    });
  }

  closeMulticastDialog = () => {
    this.setState({
      multicastDialogOpen: false,
    });
  }

  addSelectedDevicesToMulticastGroup = () => {
    this.closeMulticastDialog();

    const selected = this.state.selected;
    this.setState({
      selected: {},
    });

    for (const [key, ] of Object.entries(selected)) {
      MulticastGroupStore.addDevice(this.state.selectedMulticastGroup, key, resp => {});
    }
  }

  getRow(obj) {
    let lastseen = "n/a";
    let margin = "n/a";
    let battery = "n/a";

    if (obj.lastSeenAt !== undefined && obj.lastSeenAt !== null) {
      lastseen = moment(obj.lastSeenAt).fromNow();
    }

    if (!obj.deviceStatusExternalPowerSource && !obj.deviceStatusBatteryLevelUnavailable) {
      battery = `${obj.deviceStatusBatteryLevel}%`
    }

    if (obj.deviceStatusExternalPowerSource) {
      battery = <PowerPlug />;
    }

    if (obj.deviceStatusMargin !== undefined && obj.deviceStatusMargin !== 256) {
      margin = `${obj.deviceStatusMargin} dB`;
    }

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
        <TableCell>{lastseen}</TableCell>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${obj.devEUI}`}>{obj.name}</TableCellLink>
        <TableCell>{obj.devEUI}</TableCell>
          <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/device-profiles/${obj.deviceProfileID}`}>{obj.deviceProfileName}</TableCellLink>
        <TableCell>{margin}</TableCell>
        <TableCell>{battery}</TableCell>
      </TableRow>
    );
  }

  render() {
    const multicastGroupOptions = this.state.multicastGroups.map((mg, i) => <MenuItem value={mg.id}>{ mg.name }</MenuItem>); 

    return(
      <Grid container spacing={4}>
        <Dialog open={this.state.multicastDialogOpen} onClose={this.onMulticastDialogClose}>
          <DialogTitle>Add devices to multicast-group</DialogTitle>
          <DialogContent>
            <DialogContentText>Select the multicast-group to which the devices must be added:</DialogContentText>
            <Select fullWidth value={this.state.selectedMulticastGroup} onChange={this.onMulticastGroupSelectChange}>
              {multicastGroupOptions}
            </Select>
          </DialogContent>
          <DialogActions>
            <Button color="primary" onClick={this.closeMulticastDialog}>Cancel</Button>
            <Button color="primary" onClick={this.addSelectedDevicesToMulticastGroup}>Add</Button>
          </DialogActions>
        </Dialog>
        <DeviceAdmin organizationID={this.props.match.params.organizationID}>
          <Grid item xs={12} className={this.props.classes.buttons}>
            <Button variant="outlined" className={this.props.classes.button} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/create`}>
              <Plus className={this.props.classes.icon} />
              Create
            </Button>
            <Button variant="outlined" disabled={Object.keys(this.state.selected).length=== 0} className={this.props.classes.button} onClick={this.onSelectedMenuOpen}>
              Selected devices
            </Button>
            <Menu
              id="selected-menu"
              anchorEl={this.state.selectedMenuAnchor}
              keepMounted
              open={!!this.state.selectedMenuAnchor}
              onClose={this.onSelectedMenuClose}
            >
              <MenuItem onClick={this.openMulticastDialog}>Add to multicast group</MenuItem>
            </Menu>
          </Grid>
        </DeviceAdmin>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell></TableCell>
                <TableCell>Last seen</TableCell>
                <TableCell>Device name</TableCell>
                <TableCell>Device EUI</TableCell>
                <TableCell>Device profile</TableCell>
                <TableCell>Link margin</TableCell>
                <TableCell>Battery</TableCell>
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

export default withStyles(styles)(ListDevices);
