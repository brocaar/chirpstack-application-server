import React from "react";

import { withStyles } from "@material-ui/core/styles";
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import TextField from '@material-ui/core/TextField';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';
import FormControl from "@material-ui/core/FormControl";
import FormHelperText from "@material-ui/core/FormHelperText";

import {Controlled as CodeMirror} from "react-codemirror2";
import "codemirror/mode/javascript/javascript";

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";


const styles = {
  codeMirror: {
    zIndex: 1,
  },
};



class DeviceQueueItemForm extends FormComponent {
  constructor() {
    super();

    this.state = {
      tab: 0,
    };
  }

  onTabChange = (e, v) => {
    this.setState({
      tab: v,
    });
  }

  onCodeChange = (field, editor, data, newCode) => {
    let object = this.state.object;
    object[field] = newCode;
    this.setState({
      object: object,
    });
  }

  render() {
    if (this.state.object === undefined) {
      return null;
    }

    const codeMirrorOptions = {
      lineNumbers: true,
      mode: "javascript",
      theme: "default",
    };

    let objectCode = this.state.object.jsonObject;
    if (objectCode === "" || objectCode === undefined) {
      objectCode = `{}`
    }

    return(
      <Form
        submitLabel={this.props.submitLabel}
        onSubmit={this.onSubmit}
      >
        <TextField
          id="fPort"
          label="Port"
          margin="normal"
          value={this.state.object.fPort || ""}
          onChange={this.onChange}
          helperText="Please note that the fPort value must be > 0."
          required
          fullWidth
          type="number"
        />
        <FormControl fullWidth margin="normal">
          <FormControlLabel
            label="Confirmed downlink"
            control={
              <Checkbox
                id="confirmed"
                checked={!!this.state.object.confirmed}
                onChange={this.onChange}
                color="primary"
              />
            }
          />
        </FormControl>
        <Tabs value={this.state.tab} onChange={this.onTabChange} indicatorColor="primary">
          <Tab label="Base64 encoded" />
          <Tab label="JSON object" />
        </Tabs>
        {this.state.tab === 0 && <TextField
          id="data"
          label="Base64 encoded string"
          margin="normal"
          value={this.state.object.data || ""}
          onChange={this.onChange}
          required
          fullWidth
        />}
        {this.state.tab === 1 && <FormControl fullWidth margin="normal">
          <CodeMirror
            value={objectCode}
            className={this.props.classes.codeMirror}
            options={codeMirrorOptions}
            onBeforeChange={this.onCodeChange.bind(this, 'jsonObject')}
          />
          <FormHelperText>
            The device must be configured with a Device Profile supporting a Codec which is able to encode the given (JSON) payload.
          </FormHelperText>
        </FormControl>}
      </Form>
    );
  }
}

export default withStyles(styles)(DeviceQueueItemForm);

