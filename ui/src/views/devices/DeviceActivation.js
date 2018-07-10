import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import { withStyles } from "@material-ui/core/styles";
import Typograhy from "@material-ui/core/Typography";
import Card from '@material-ui/core/Card';
import CardContent from "@material-ui/core/CardContent";
import TextField from "@material-ui/core/TextField";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import DeviceStore from "../../stores/DeviceStore";
import theme from "../../theme";


const styles = {
  link: {
    color: theme.palette.primary.main,
  },
};


class DeviceActivationForm extends FormComponent {
  constructor() {
    super();
    this.getRandomDevAddr = this.getRandomDevAddr.bind(this);
  }

  getRandomDevAddr(e) {
    e.preventDefault();

    if (this.props.disabled) {
      return;
    }

    DeviceStore.getRandomDevAddr(this.props.match.params.devEUI, resp => {
      let object = this.state.object;
      object.devAddr = resp.devAddr;
      this.setState({
        object: object,
      });
    });
  }

  getRandomKey(field, e) {
    e.preventDefault();

    if (this.props.disabled) {
      return;
    }

    let object = this.state.object;
    let key = "";
    const possible = 'abcdef0123456789';

    for(let i = 0; i < 32; i++){
      key += possible.charAt(Math.floor(Math.random() * possible.length));
    }

    object[field] = key;

    if (field === "nwkSEncKey" && this.props.deviceProfile.macVersion.startsWith("1.0")) {
      object["sNwkSIntKey"] = key;
      object["fNwkSIntKey"] = key;
    }

    this.setState({
      object: object,
    });
  }

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
          id="devAddr"
          label="Device address"
          helperText={<span><a href="#random" onClick={this.getRandomDevAddr} className={this.props.classes.link}>Generate random address</a>.</span>}
          margin="normal"
          value={this.state.object.devAddr || ""}
          placeholder="00000000"
          onChange={this.onChange}
          inputProps={{
            pattern: "[A-Fa-f0-9]{8}",
          }}
          disabled={this.props.disabled}
          fullWidth
          required
        />
        <TextField
          id="nwkSEncKey"
          label="Network session encryption key"
          helperText={<span><a href="#random" onClick={this.getRandomKey.bind(this, "nwkSEncKey")} className={this.props.classes.link}>Generate random key</a>. For LoRaWAN 1.0 devices, this value holds the NwkSKey.</span>}
          margin="normal"
          value={this.state.object.nwkSEncKey || ""}
          placeholder="00000000000000000000000000000000"
          onChange={this.onChange}
          inputProps={{
            pattern: "[A-Fa-f0-9]{32}",
          }}
          disabled={this.props.disabled}
          fullWidth
          required
        />
        <TextField
          id="sNwkSIntKey"
          label="Serving network session integrity key"
          margin="normal"
          value={this.state.object.sNwkSIntKey || ""}
          placeholder="00000000000000000000000000000000"
          helperText={<span><a href="#random" onClick={this.getRandomKey.bind(this, "sNwkSIntKey")} className={this.props.classes.link}>Generate random key</a>. For LoRaWAN 1.0 devices, this value holds the NwkSKey.</span>}
          onChange={this.onChange}
          inputProps={{
            pattern: "[A-Fa-f0-9]{32}",
          }}
          disabled={this.props.disabled}
          fullWidth
          required
        />
        <TextField
          id="fNwkSIntKey"
          label="Forwarding network session integrity key"
          margin="normal"
          value={this.state.object.fNwkSIntKey || ""}
          placeholder="00000000000000000000000000000000"
          helperText={<span><a href="#random" onClick={this.getRandomKey.bind(this, "fNwkSIntKey")} className={this.props.classes.link}>Generate random key</a>. For LoRaWAN 1.0 devices, this value holds the NwkSKey.</span>}
          onChange={this.onChange}
          inputProps={{
            pattern: "[A-Fa-f0-9]{32}",
          }}
          disabled={this.props.disabled}
          fullWidth
          required
        />
        <TextField
          id="appSKey"
          label="Application session key"
          margin="normal"
          value={this.state.object.appSKey || ""}
          placeholder="00000000000000000000000000000000"
          helperText={<span><a href="#random" onClick={this.getRandomKey.bind(this, "appSKey")} className={this.props.classes.link}>Generate random key</a>.</span>}
          onChange={this.onChange}
          inputProps={{
            pattern: "[A-Fa-f0-9]{32}",
          }}
          disabled={this.props.disabled}
          required
          fullWidth
        />
        <TextField
          id="fCntUp"
          label="Uplink frame-counter"
          margin="normal"
          type="number"
          value={this.state.object.fCntUp || 0}
          onChange={this.onChange}
          disabled={this.props.disabled}
          required
          fullWidth
        />
        <TextField
          id="nFCntDown"
          label="Downlink frame-counter (network)"
          margin="normal"
          type="number"
          value={this.state.object.nFCntDown || 0}
          onChange={this.onChange}
          disabled={this.props.disabled}
          required
          fullWidth
        />
        <TextField
          id="aFCntDown"
          label="Downlink frame-counter (application)"
          margin="normal"
          helperText="This frame-counter is only used for LoRaWAN 1.1+ devices."
          type="number"
          value={this.state.object.aFCntDown || 0}
          onChange={this.onChange}
          disabled={this.props.disabled}
          required
          fullWidth
        />
      </Form>
    );
  }
}


DeviceActivationForm = withStyles(styles)(DeviceActivationForm);


class DeviceActivation extends Component {
  constructor() {
    super();
    this.state = {};
    this.onSubmit = this.onSubmit.bind(this);
  }
  
  componentDidMount() {
    DeviceStore.getActivation(this.props.match.params.devEUI, resp => {
      this.setState({
        deviceActivation: resp,
      });
    });
  }

  onSubmit(deviceActivation) {
    let act = deviceActivation;
    act.devEUI = this.props.match.params.devEUI;
    DeviceStore.activate(act, resp => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
    });
  }

  render() {
    let object;
    if (this.state.deviceActivation !== undefined) {
      object = this.state.deviceActivation.deviceActivation;
    }

    let submitLabel = null;
    if (!this.props.deviceProfile.supportsJoin) {
      submitLabel = "(Re)activate device";
    }

    let showForm = false;
    if (!this.props.deviceProfile.supportsJoin || (this.props.deviceProfile.supportsJoin && object)) {
      showForm = true;
    }

    return(
      <Card>
        <CardContent>
          {showForm && <DeviceActivationForm
            submitLabel={submitLabel}
            object={object}
            onSubmit={this.onSubmit}
            disabled={this.props.deviceProfile.supportsJoin}
            match={this.props.match}
            deviceProfile={this.props.deviceProfile}
          />}
          {!showForm && <Typograhy variant="body1">
            This device has not (yet) been activated.
          </Typograhy>}
        </CardContent>
      </Card>
    );
  }
}

export default withRouter(DeviceActivation);
