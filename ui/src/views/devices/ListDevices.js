import React, { Component } from "react";
import { Link } from "react-router-dom";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Button from '@material-ui/core/Button';
import Radio from "@material-ui/core/Radio";
import ArrowDropDown from '@material-ui/icons/ArrowDropDown';
import ArrowDropUp from '@material-ui/icons/ArrowDropUp';

import moment from "moment";
import Plus from "mdi-material-ui/Plus";
import PowerPlug from "mdi-material-ui/PowerPlug";

import TableCellLink from "../../components/TableCellLink";
import DataTable from "../../components/DataTable";
import Admin from "../../components/Admin";
import DeviceStore from "../../stores/DeviceStore";
import theme from "../../theme";


const styles = {
  buttons: {
    textAlign: "right",
  },
  button: {
    marginLeft: 2 * theme.spacing.unit,
  },
  icon: {
    marginRight: theme.spacing.unit,
  },
  arrowUpIcon:{
    paddingRight: theme.spacing.unit,
  },
  arrowDownIcon:{
    paddingLeft: 0,
  }
};


class ListDevices extends Component {
  constructor() {
    super();

    this.state = {
      order: 'asc',
      orderBy: 'd.name'
    }	  

    this.getPage = this.getPage.bind(this);
    this.getRow = this.getRow.bind(this);
  }

  handleAsc = event => {
    this.setState({ 
	orderBy: event.target.value, 
	order: 'asc'
	});
  };

 handleDesc = event => {
    this.setState({ 
	orderBy: event.target.value, 
    	order: 'desc'
    });
	
};

  getPage(limit, offset, callbackFunc) {
    DeviceStore.list({
      applicationID: this.props.match.params.applicationID,
      limit: limit,
      offset: offset,
      order: this.state.order,
      orderBy: this.state.orderBy
    }, callbackFunc);
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
      <TableRow key={obj.devEUI}>
        <TableCell>{lastseen}</TableCell>
        <TableCellLink to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${obj.devEUI}`}>{obj.name}</TableCellLink>
        <TableCell>{obj.devEUI}</TableCell>
        <TableCell>{margin}</TableCell>
        <TableCell>{battery}</TableCell>
      </TableRow>
    );
  }

  render() {
    return(
      <Grid container spacing={24}>
        <Admin organizationID={this.props.match.params.organizationID}>
          <Grid item xs={12} className={this.props.classes.buttons}>
            <Button variant="outlined" className={this.props.classes.button} component={Link} to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/create`}>
              <Plus className={this.props.classes.icon} />
              Create
            </Button>
          </Grid>
        </Admin>
        <Grid item xs={12}>
          <DataTable
            header={
              <TableRow>
                <TableCell>Last seen
			<Radio
		            checked={this.state.orderBy === "d.last_seen_at" && this.state.order === 'asc'}
		            onChange={this.handleAsc}
		            value="d.last_seen_at"
            		    name="radio-button-demo"
            		    aria-label="last_seen_at"
            		    icon={<ArrowDropUp fontSize="small" />}
                            checkedIcon={ <ArrowDropUp fontSize="small" /> }
		            className={this.props.classes.arrowUpIcon}
          		/>
                        <Radio
                            checked={this.state.orderBy === "d.last_seen_at" && this.state.order === 'desc'}
                            onChange={this.handleDesc}
                            value="d.last_seen_at"
                            name="radio-button-demo"
                            aria-label="last_seen_at"
                            icon={<ArrowDropDown fontSize="small" />}
                            checkedIcon={ <ArrowDropDown fontSize="small" /> }
			    className={this.props.classes.arrowDownIcon}
                        />


		</TableCell>
                <TableCell>Device name
			<Radio
            			checked={this.state.orderBy === "d.name" && this.state.order === 'asc'}
            			onChange={this.handleAsc}
            			value="d.name"
            			name="radio-button-demo"
            			aria-label="name"
                                icon={<ArrowDropUp fontSize="small" />}
                                checkedIcon={ <ArrowDropUp fontSize="small" /> }
				className={this.props.classes.arrowUpIcon}
          		/>
                        <Radio
                                checked={this.state.orderBy === "d.name" && this.state.order === 'desc'}
                                onChange={this.handleDesc}
                                value="d.name"
                                name="radio-button-demo"
                                aria-label="name"
                                icon={<ArrowDropDown fontSize="small" />}
                                checkedIcon={ <ArrowDropDown fontSize="small" /> }
                       		className={this.props.classes.arrowDownIcon}
			 />

		</TableCell>
                <TableCell>Device EUI
			<Radio
            			checked={this.state.orderBy === "d.dev_eui" && this.state.order === 'asc'}
            			onChange={this.handleAsc}
            			value="d.dev_eui"
            			name="radio-button-demo"
            			aria-label="devEUI"
                                icon={<ArrowDropUp fontSize="small" />}
                                checkedIcon={ <ArrowDropUp fontSize="small" /> }
				className={this.props.classes.arrowUpIcon}
          		/>
                        <Radio
                                checked={this.state.orderBy === "d.dev_eui" && this.state.order === 'desc'}
                                onChange={this.handleDesc}
                                value="d.dev_eui"
                                name="radio-button-demo"
                                aria-label="devEui"
                                icon={<ArrowDropDown fontSize="small" />}
                                checkedIcon={ <ArrowDropDown fontSize="small" /> }
                        	className={this.props.classes.arrowDownIcon}
			/>

		</TableCell>
                <TableCell>Link margin
			<Radio
            			checked={this.state.orderBy === "d.device_status_margin" && this.state.order === 'asc'}
            			onChange={this.handleAsc}
            			value="d.device_status_margin"
            			name="radio-button-demo"
            			aria-label="deviceStatusMargin"
                                icon={<ArrowDropUp fontSize="small" />}
                                checkedIcon={ <ArrowDropUp fontSize="small" /> }
				className={this.props.classes.arrowUpIcon}
          		/>
                        <Radio
                                checked={this.state.orderBy === "d.device_status_margin" && this.state.order === 'desc'}
                                onChange={this.handleDesc}
                                value="d.device_status_margin"
                                name="radio-button-demo"
                                aria-label="deviceStatusMargin"
                                icon={<ArrowDropDown fontSize="small" />}
                                checkedIcon={ <ArrowDropDown fontSize="small" /> }
                        	className={this.props.classes.arrowDownIcon}
			/>

		</TableCell>
                <TableCell>Battery
			<Radio
        			checked={this.state.orderBy === "d.device_status_battery" && this.state.order === 'asc'}
            			onChange={this.handleAsc}
            			value="d.device_status_battery"
            			name="radio-button-demo"
            			aria-label="deviceStatusBatteryLevel"
                                icon={<ArrowDropUp fontSize="small" />}
                                checkedIcon={ <ArrowDropUp fontSize="small" /> }
				className={this.props.classes.arrowUpIcon}
          		/>
                        <Radio
                                checked={this.state.orderBy === "d.device_status_battery" && this.state.order === 'desc'}
                                onChange={this.handleDesc}
                                value="d.device_status_battery"
                                name="radio-button-demo"
                                aria-label="deviceStatusBattery"
				icon={<ArrowDropDown fontSize="small" />}
                                checkedIcon={ <ArrowDropDown fontSize="small" /> }                                
                        	className={this.props.classes.arrowDownIcon}
			/>

		</TableCell>
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
