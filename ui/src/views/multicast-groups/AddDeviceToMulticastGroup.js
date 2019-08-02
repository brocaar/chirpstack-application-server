import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";
import FormControl from "@material-ui/core/FormControl";
import FormLabel from "@material-ui/core/FormLabel";
import FormHelperText from "@material-ui/core/FormHelperText";

import Form from "../../components/Form";
import TitleBar from "../../components/TitleBar";
import TitleBarTitle from "../../components/TitleBarTitle";
import FormComponent from "../../classes/FormComponent";
import MulticastGroupStore from "../../stores/MulticastGroupStore";
import DeviceStore from "../../stores/DeviceStore";
import AutocompleteSelect from "../../components/AutocompleteSelect";


const styles = {
  card: {
    overflow: "visible",
  },
  formLabel: {
    fontSize: 12,
  },
};


class AddDeviceForm extends FormComponent {
  constructor() {
    super();

    this.getDeviceOption = this.getDeviceOption.bind(this);
    this.getDeviceOptions = this.getDeviceOptions.bind(this);
  }

  getDeviceOption(devEUI, callbackFunc) {
    DeviceStore.get(devEUI, resp => {
      callbackFunc({label: `${resp.device.name} (${resp.device.devEUI})`, value: resp.device.devEUI});
    });
  }

  getDeviceOptions(search, callbackFunc) {
    DeviceStore.list({serviceProfileID: this.props.serviceProfileID, search: search, limit: 10}, resp => {
      const options = resp.result.map((d, i) => {return {label: `${d.name} (${d.devEUI})`, value: d.devEUI}});
      callbackFunc(options);
    });
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <FormControl fullWidth margin="normal">
          <FormLabel className={this.props.classes.FormLabel} required>Device</FormLabel>
          <AutocompleteSelect
            id="devEUI"
            label="Select device"
            value={this.state.object.devEUI || ""}
            onChange={this.onChange}
            getOption={this.getDeviceOption}
            getOptions={this.getDeviceOptions}
            margin="none"
          />
          <FormHelperText>Search on device by it's name or device EUI. Only devices part of the same service-profile as the multicast-group can be added (and are visible).</FormHelperText>
        </FormControl>
      </Form>
    );
  }
}

AddDeviceForm = withStyles(styles)(AddDeviceForm);


class AddDeviceToMulticastGroup extends Component {
  constructor() {
    super();
    this.state = {};
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    MulticastGroupStore.get(this.props.match.params.multicastGroupID, resp => {
      this.setState({
        multicastGroup: resp.multicastGroup,
      });
    });
  }

  onSubmit(device) {
    MulticastGroupStore.addDevice(this.props.match.params.multicastGroupID, device.devEUI, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/multicast-groups/${this.props.match.params.multicastGroupID}`);
    });
  }

  render() {
    if (this.state.multicastGroup === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <TitleBar>
          <TitleBarTitle title="Multicast-groups" to={`/organizations/${this.props.match.params.organizationID}/multicast-groups`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title={this.state.multicastGroup.name} to={`/organizations/${this.props.match.params.organizationID}/multicast-groups/${this.state.multicastGroup.id}`} />
          <TitleBarTitle title="/" />
          <TitleBarTitle title="Add device" />
        </TitleBar>

        <Grid item xs={12}>
          <Card className={this.props.classes.card}>
            <CardContent>
              <AddDeviceForm submitLabel="Add device" onSubmit={this.onSubmit} serviceProfileID={this.state.multicastGroup.serviceProfileID} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withStyles(styles)(withRouter(AddDeviceToMulticastGroup))
