import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import AESKeyField from "../../components/AESKeyField";
import DeviceStore from "../../stores/DeviceStore";


class LW11DeviceKeysForm extends FormComponent {
  render() {
    let object = {};
    if (this.props.object !== undefined) {
      object = this.props.object;
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <AESKeyField
          id="nwkKey"
          label="Network key (LoRaWAN 1.1)"
          helperText="For LoRaWAN 1.1 devices. In case your device does not support LoRaWAN 1.1, update the device-profile first."
          onChange={this.onChange}
          value={object.nwkKey || ""}
          margin="normal"
          fullWidth
          required
          random
        />
        <AESKeyField
          id="appKey"
          label="Application key (LoRaWAN 1.1)"
          helperText="For LoRaWAN 1.1 devices. In case your device does not support LoRaWAN 1.1, update the device-profile first."
          onChange={this.onChange}
          value={object.appKey || ""}
          margin="normal"
          fullWidth
          required
          random
        />
      </Form>
    );
  }
}

class LW10DeviceKeysForm extends FormComponent {
  render() {
    let object = {};
    if (this.props.object !== undefined) {
      object = this.props.object;
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <AESKeyField
          id="nwkKey"
          label="Application key"
          helperText="For LoRaWAN 1.0 devices. In case your device supports LoRaWAN 1.1, update the device-profile first."
          onChange={this.onChange}
          value={object.nwkKey || ""}
          margin="normal"
          fullWidth
          required
          random
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
      if (resp === null) {
        this.setState({
          deviceKeys: {
            deviceKeys: {},
          },
        });
      } else {
        this.setState({
          update: true,
          deviceKeys: resp,
        });
      }
    });
  }

  onSubmit(deviceKeys) {
    if (this.state.update) {
      DeviceStore.updateKeys(deviceKeys, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}`);
      });
    } else {
      let keys = deviceKeys;
      keys.devEUI = this.props.match.params.devEUI;
      DeviceStore.createKeys(keys, resp => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/devices/${this.props.match.params.devEUI}`);
      });
    }
  }

  render() {
    if (this.state.deviceKeys === undefined) {
      return null;
    }

    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              {this.props.deviceProfile.macVersion.startsWith("1.0") && <LW10DeviceKeysForm
                submitLabel="Set device-keys"
                onSubmit={this.onSubmit}
                object={this.state.deviceKeys.deviceKeys}
              />}
              {this.props.deviceProfile.macVersion.startsWith("1.1") && <LW11DeviceKeysForm
                submitLabel="Set device-keys"
                onSubmit={this.onSubmit}
                object={this.state.deviceKeys.deviceKeys}
              />}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default withRouter(DeviceKeys);
