import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";
import TextField from "@material-ui/core/TextField";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import DeviceStore from "../../stores/DeviceStore";


class DeviceKeysForm extends FormComponent {
  render() {
    if (this.state.object === undefined) {
      return(<div></div>);
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <TextField
          id="nwkKey"
          label="Network key"
          helperText="For LoRaWAN 1.0 devices, this is the only key you need to set (in LoRaWAN 1.0 this used to be the application-key)."
          placeholder="00000000000000000000000000000000"
          inputProps={{
            pattern: "[A-Fa-f0-9]{32}",
          }}
          onChange={this.onChange}
          value={this.state.object.nwkKey || ""}
          margin="normal"
          fullWidth
          required
        />
        <TextField
          id="appKey"
          label="Application key"
          helperText="Leave this blank for LoRaWAN 1.0 devices."
          placeholder="00000000000000000000000000000000"
          inputProps={{
            pattern: "[A-Fa-f0-9]{32}",
          }}
          onChange={this.onChange}
          value={this.state.object.appKey || ""}
          margin="normal"
          fullWidth
        />
      </Form>
    );
  }
}


class DeviceKeys extends Component {
  constructor() {
    super();
    this.state = {
      update: false,
    };
    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    DeviceStore.getKeys(this.props.match.params.devEUI, resp => {
      this.setState({
        update: true,
        deviceKeys: resp,
      });
    });
  }

  onSubmit(deviceKeys) {
    if (this.state.update) {
      DeviceStore.updateKeys(deviceKeys, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
      });
    } else {
      let keys = deviceKeys;
      keys.devEUI = this.props.match.params.devEUI;
      DeviceStore.createKeys(keys, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
      });
    }
  }

  render() {
    let object;
    if (this.state.deviceKeys !== undefined) {
      object = this.state.deviceKeys.deviceKeys;
    }

    return(
      <Grid container spacing={24}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <DeviceKeysForm
                submitLabel="Set device-keys"
                onSubmit={this.onSubmit}
                object={object}
              />
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(DeviceKeys);
