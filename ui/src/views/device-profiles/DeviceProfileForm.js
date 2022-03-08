import React from "react";

import { withStyles } from "@material-ui/core/styles";
import TextField from '@material-ui/core/TextField';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";

import {Controlled as CodeMirror} from "react-codemirror2";
import "codemirror/mode/javascript/javascript";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";
import KVForm from "../../components/KVForm";
import DurationField from "../../components/DurationField";
import AutocompleteSelect from "../../components/AutocompleteSelect";
import NetworkServerStore from "../../stores/NetworkServerStore";
import { FormLabel } from "../../../node_modules/@material-ui/core";


const styles = {
  formLabel: {
    fontSize: 12,
  },
  codeMirror: {
    zIndex: 1,
  },
};


class DeviceProfileForm extends FormComponent {
  constructor() {
    super();
    this.state = {
      tab: 0,
      tags: [],
    };

    this.onTabChange = this.onTabChange.bind(this);
    this.getNetworkServerOptions = this.getNetworkServerOptions.bind(this);
    this.getMACVersionOptions = this.getMACVersionOptions.bind(this);
    this.getRegParamsOptions = this.getRegParamsOptions.bind(this);
    this.getPingSlotPeriodOptions = this.getPingSlotPeriodOptions.bind(this);
    this.getPayloadCodecOptions = this.getPayloadCodecOptions.bind(this);
    this.onCodeChange = this.onCodeChange.bind(this);
  }

  componentDidMount() {
    super.componentDidMount();
    this.setKVArray(this.props.object || {});
  }

  componentDidUpdate(prevProps) {
    super.componentDidUpdate(prevProps);

    if (prevProps.object !== this.props.object) {
      this.setKVArray(this.props.object || {});
    }
  }

  getNetworkServerOptions(search, callbackFunc) {
    NetworkServerStore.list(this.props.match.params.organizationID, 999, 0, resp => {
      const options = resp.result.map((ns, i) => {return {label: ns.name, value: ns.id}});
      callbackFunc(options);
    });
  }

  getMACVersionOptions(search, callbackFunc) {
    const macVersionOptions = [
      {value: "1.0.0", label: "1.0.0"},
      {value: "1.0.1", label: "1.0.1"},
      {value: "1.0.2", label: "1.0.2"},
      {value: "1.0.3", label: "1.0.3"},
      {value: "1.0.4", label: "1.0.4"},
      {value: "1.1.0", label: "1.1.0"},
    ];

    callbackFunc(macVersionOptions);
  }

  getRegParamsOptions(search, callbackFunc) {
    const regParamsOptions = [
      {value: "A", label: "A"},
      {value: "B", label: "B"},
      {value: "RP002-1.0.0", label: "RP002-1.0.0"},
      {value: "RP002-1.0.1", label: "RP002-1.0.1"},
      {value: "RP002-1.0.2", label: "RP002-1.0.2"},
      {value: "RP002-1.0.3", label: "RP002-1.0.3"},
    ];

    callbackFunc(regParamsOptions);
  }

  getPingSlotPeriodOptions(search, callbackFunc) {
    const pingSlotPeriodOptions = [
      {value: 32 * 1, label: "every second"},
      {value: 32 * 2, label: "every 2 seconds"},
      {value: 32 * 4, label: "every 4 seconds"},
      {value: 32 * 8, label: "every 8 seconds"},
      {value: 32 * 16, label: "every 16 seconds"},
      {value: 32 * 32, label: "every 32 seconds"},
      {value: 32 * 64, label: "every 64 seconds"},
      {value: 32 * 128, label: "every 128 seconds"},
    ];

    callbackFunc(pingSlotPeriodOptions);
  }

  getPayloadCodecOptions(search, callbackFunc) {
    const payloadCodecOptions = [
      {value: "", label: "None"},
      {value: "CAYENNE_LPP", label: "Cayenne LPP"},
      {value: "CUSTOM_JS", label: "Custom JavaScript codec functions"},
    ];

    callbackFunc(payloadCodecOptions);
  }

  getADRAlgorithmsOptions = (search, callbackFunc) => {
    if (this.state.object.networkServerID === undefined) {
      callbackFunc([]);
    } else {
      NetworkServerStore.getADRAlgorithms(this.state.object.networkServerID, resp => {
        const options = resp.adrAlgorithms.map((adr, i) => {return {value: adr.id, label: adr.name}});
        callbackFunc(options);
      })
    }
  }

  onCodeChange(field, editor, data, newCode) {
    let object = this.state.object;
    object[field] = newCode;
    this.setState({
      object: object,
    });
  }

  onTabChange(e, v) {
    this.setState({
      tab: v,
    });
  }

  onChange(e) {
    super.onChange(e);
    if (e.target.id === "factoryPresetFreqsStr") {
      let object = this.state.object;
      let freqsStr = e.target.value.split(",");
      object["factoryPresetFreqs"] = freqsStr.map((f, i) => parseInt(f, 10));
      this.setState({
        object: object,
      });
    }
  }

  setKVArray = (props) => {
    let tags = [];

    if (props.tags !== undefined) {
      for (let key in props.tags) {
        tags.push({key: key, value: props.tags[key]});
      }
    }

    this.setState({
      tags: tags,
    });
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    let factoryPresetFreqsStr = "";
    if (this.state.object.factoryPresetFreqsStr !== undefined) {
      factoryPresetFreqsStr = this.state.object.factoryPresetFreqsStr;
    } else if (this.state.object.factoryPresetFreqs !== undefined) {
      factoryPresetFreqsStr = this.state.object.factoryPresetFreqs.join(", ");
    }

    const codeMirrorOptions = {
      lineNumbers: true,
      mode: "javascript",
      theme: "default",
    };
    
    let payloadEncoderScript = this.state.object.payloadEncoderScript;
    let payloadDecoderScript = this.state.object.payloadDecoderScript;

    if (payloadEncoderScript === "" || payloadEncoderScript === undefined) {
      payloadEncoderScript = `// Encode encodes the given object into an array of bytes.
//  - fPort contains the LoRaWAN fPort number
//  - obj is an object, e.g. {"temperature": 22.5}
//  - variables contains the device variables e.g. {"calibration": "3.5"} (both the key / value are of type string)
// The function must return an array of bytes, e.g. [225, 230, 255, 0]
function Encode(fPort, obj, variables) {
  return [];
}`;
    }

    if (payloadDecoderScript === "" || payloadDecoderScript === undefined) {
      payloadDecoderScript = `// Decode decodes an array of bytes into an object.
//  - fPort contains the LoRaWAN fPort number
//  - bytes is an array of bytes, e.g. [225, 230, 255, 0]
//  - variables contains the device variables e.g. {"calibration": "3.5"} (both the key / value are of type string)
// The function must return an object, e.g. {"temperature": 22.5}
function Decode(fPort, bytes, variables) {
  return {};
}`;
    }

    const tags = this.state.tags.map((obj, i) => <KVForm key={i} index={i} object={obj} onChange={this.onChangeKV("tags")} onDelete={this.onDeleteKV("tags")} />);


    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
        disabled={this.props.disabled}
      >
        <Tabs value={this.state.tab} onChange={this.onTabChange} indicatorColor="primary">
          <Tab label="General" />
          <Tab label="Join (OTAA / ABP)" />
          <Tab label="Class-B" />
          <Tab label="Class-C" />
          <Tab label="Codec" />
          <Tab label="Tags" />
        </Tabs>

        {this.state.tab === 0 && <div>
          <TextField
            id="name"
            label="Device-profile name"
            margin="normal"
            value={this.state.object.name || ""}
            onChange={this.onChange}
            helperText="A name to identify the device-profile."
            required
            fullWidth
          />
          {!this.props.update && <FormControl fullWidth margin="normal">
            <FormLabel className={this.props.classes.formLabel} required>Network-server</FormLabel>
            <AutocompleteSelect
              id="networkServerID"
              label="Select network-server"
              value={this.state.object.networkServerID || ""}
              onChange={this.onChange}
              getOptions={this.getNetworkServerOptions}
            />
            <FormHelperText>
              The network-server on which this device-profile will be provisioned. After creating the device-profile, this value can't be changed.
            </FormHelperText>
          </FormControl>}
          <FormControl fullWidth margin="normal">
            <FormLabel className={this.props.classes.formLabel} required>LoRaWAN MAC version</FormLabel>
            <AutocompleteSelect
              id="macVersion"
              label="Select LoRaWAN MAC version"
              value={this.state.object.macVersion || ""}
              onChange={this.onChange}
              getOptions={this.getMACVersionOptions}
            />
            <FormHelperText>
              The LoRaWAN MAC version supported by the device.
            </FormHelperText>
          </FormControl>
          <FormControl fullWidth margin="normal">
            <FormLabel className={this.props.classes.formLabel} required>LoRaWAN Regional Parameters revision</FormLabel>
            <AutocompleteSelect
              id="regParamsRevision"
              label="Select LoRaWAN Regional Parameters revision"
              value={this.state.object.regParamsRevision || ""}
              onChange={this.onChange}
              getOptions={this.getRegParamsOptions}
            />
            <FormHelperText>
                Revision of the Regional Parameters specification supported by the device.
            </FormHelperText>
          </FormControl>
          <FormControl fullWidth margin="normal">
            <FormLabel className={this.props.classes.formLabel} required>ADR algorithm</FormLabel>
            <AutocompleteSelect
              id="adrAlgorithmID"
              label="Select ADR algorithm"
              value={this.state.object.adrAlgorithmID || ""}
              onChange={this.onChange}
              getOptions={this.getADRAlgorithmsOptions}
              triggerReload={this.state.object.networkServerID || ""}
            />
            <FormHelperText>
                The ADR algorithm that will be used for controlling the device data-rate.
            </FormHelperText>
          </FormControl>
          <TextField
            id="maxEIRP"
            label="Max EIRP"
            type="number"
            margin="normal"
            value={this.state.object.maxEIRP || 0}
            onChange={this.onChange}
            helperText="Maximum EIRP supported by the device."
            required
            fullWidth
          />
          <DurationField
            id="uplinkInterval"
            label="Uplink interval (seconds)"
            helperText="The expected interval in seconds in which the device sends uplink messages. This is used to determine if a device is active or inactive."
            value={this.state.object.uplinkInterval}
            onChange={this.onChange}
          />
        </div>}

        {this.state.tab === 1 && <div>
          <FormControl fullWidth margin="normal">
            <FormControlLabel
              label="Device supports OTAA"
              control={
                <Checkbox
                  id="supportsJoin"
                  checked={!!this.state.object.supportsJoin}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
          </FormControl>
          {!this.state.object.supportsJoin && <TextField
            id="rxDelay1"
            label="RX1 delay"
            type="number"
            margin="normal"
            value={this.state.object.rxDelay1 || 0}
            onChange={this.onChange}
            helperText="RX1 delay (valid values are 0 - 15)."
            required
            fullWidth
          />}
          {!this.state.object.supportsJoin && <TextField
            id="rxDROffset1"
            label="RX1 data-rate offset"
            type="number"
            margin="normal"
            value={this.state.object.rxDROffset1 || 0}
            onChange={this.onChange}
            helperText="Please refer the LoRaWAN Regional Parameters specification for valid values."
            required
            fullWidth
          />}
          {!this.state.object.supportsJoin && <TextField
            id="rxDataRate2"
            label="RX2 data-rate"
            type="number"
            margin="normal"
            value={this.state.object.rxDataRate2 || 0}
            onChange={this.onChange}
            helperText="Please refer the LoRaWAN Regional Parameters specification for valid values."
            required
            fullWidth
          />}
          {!this.state.object.supportsJoin && <TextField
            id="rxFreq2"
            label="RX2 channel frequency (Hz)"
            type="number"
            margin="normal"
            value={this.state.object.rxFreq2 || 0}
            onChange={this.onChange}
            required
            fullWidth
          />}
          {!this.state.object.supportsJoin && <TextField
            id="factoryPresetFreqsStr"
            label="Factory-preset frequencies (Hz)"
            margin="normal"
            value={factoryPresetFreqsStr}
            onChange={this.onChange}
            helperText="List of factory-preset frequencies (Hz), comma separated."
            required
            fullWidth
          />}
        </div>}

        {this.state.tab === 2 && <div>
          <FormControl fullWidth margin="normal">
            <FormControlLabel
              label="Device supports Class-B"
              control={
                <Checkbox
                  id="supportsClassB"
                  checked={!!this.state.object.supportsClassB}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
          </FormControl>

          {this.state.object.supportsClassB && <TextField
            id="classBTimeout"
            label="Class-B confirmed downlink timeout"
            type="number"
            margin="normal"
            value={this.state.object.classBTimeout || 0}
            onChange={this.onChange}
            helperText="Class-B timeout (in seconds) for confirmed downlink transmissions."
            required
            fullWidth
          />}
          {this.state.object.supportsClassB && <FormControl
              fullWidth
              margin="normal"
            >
              <FormLabel className={this.props.classes.formLabel} required>Class-B ping-slot periodicity</FormLabel>
              <AutocompleteSelect
                id="pingSlotPeriod"
                label="Select Class-B ping-slot periodicity"
                value={this.state.object.pingSlotPeriod || ""}
                onChange={this.onChange}
                getOptions={this.getPingSlotPeriodOptions}
              />
              <FormHelperText>Class-B ping-slot periodicity.</FormHelperText>
          </FormControl>}
          {this.state.object.supportsClassB && <TextField
            id="pingSlotDR"
            label="Class-B ping-slot data-rate"
            type="number"
            margin="normal"
            value={this.state.object.pingSlotDR || 0}
            onChange={this.onChange}
            required
            fullWidth
          />}
          {this.state.object.supportsClassB && <TextField
            id="pingSlotFreq"
            label="Class-B ping-slot frequency (Hz)"
            type="number"
            margin="normal"
            value={this.state.object.pingSlotFreq || 0}
            onChange={this.onChange}
            required
            fullWidth
          />}
        </div>}

        {this.state.tab === 3 && <div>
          <FormControl fullWidth margin="normal">
            <FormControlLabel
              label="Device supports Class-C"
              control={
                <Checkbox
                  id="supportsClassC"
                  checked={!!this.state.object.supportsClassC}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
            <FormHelperText>Select this option when the device will operate as Class-C device immediately after activation. In case it sends a DeviceModeInd mac-command when it changes to Class-C, do not select this option.</FormHelperText>
          </FormControl>

          <TextField
            id="classCTimeout"
            label="Class-C confirmed downlink timeout"
            type="number"
            margin="normal"
            value={this.state.object.classCTimeout || 0}
            onChange={this.onChange}
            helperText="Class-C timeout (in seconds) for confirmed downlink transmissions."
            required
            fullWidth
          />
        </div>}

        {this.state.tab === 4 && <div>
          <FormControl fullWidth margin="normal">
            <FormLabel className={this.props.classes.formLabel}>Payload codec</FormLabel>
            <AutocompleteSelect
              id="payloadCodec"
              label="Select payload codec"
              value={this.state.object.payloadCodec || ""}
              onChange={this.onChange}
              getOptions={this.getPayloadCodecOptions}
            />
            <FormHelperText>
              By defining a payload codec, ChirpStack Application Server can encode and decode the binary device payload for you.
            </FormHelperText>
          </FormControl>

          {this.state.object.payloadCodec === "CUSTOM_JS" && <FormControl fullWidth margin="normal">
            <CodeMirror
              value={payloadDecoderScript}
              options={codeMirrorOptions}
              onBeforeChange={this.onCodeChange.bind(this, 'payloadDecoderScript')}
              className={this.props.classes.codeMirror}
            />
            <FormHelperText>
              The function must have the signature <strong>function Decode(fPort, bytes)</strong> and must return an object.
              ChirpStack Application Server will convert this object to JSON.
            </FormHelperText>
          </FormControl>}
          {this.state.object.payloadCodec === "CUSTOM_JS" && <FormControl fullWidth margin="normal">
            <CodeMirror
              value={payloadEncoderScript}
              options={codeMirrorOptions}
              onBeforeChange={this.onCodeChange.bind(this, 'payloadEncoderScript')}
              className={this.props.classes.codeMirror}
            />
            <FormHelperText>
              The function must have the signature <strong>function Encode(fPort, obj)</strong> and must return an array
              of bytes.
            </FormHelperText>
          </FormControl>}
        </div>}

        {this.state.tab === 5 && <div>
          <FormControl fullWidth margin="normal">
            <Typography variant="body1">
              Tags can be used to store additional key/value data.
            </Typography>
            {tags}
          </FormControl>
          <Button variant="outlined" onClick={this.addKV("tags")}>Add tag</Button>
        </div>}
      </Form>
    );
  }
}

export default withStyles(styles)(DeviceProfileForm);
