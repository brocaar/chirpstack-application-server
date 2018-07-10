import React from "react";

import TextField from '@material-ui/core/TextField';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormGroup from "@material-ui/core/FormGroup";
import FormHelperText from '@material-ui/core/FormHelperText';
import Checkbox from '@material-ui/core/Checkbox';

import FormControl from "../../components/FormControl";
import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";



class OrganizationForm extends FormComponent {
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
          id="name"
          label="Organization name"
          helperText="The name may only contain words, numbers and dashes."
          margin="normal"
          value={this.state.object.name || ""}
          onChange={this.onChange}
          inputProps={{
            pattern: "[\\w-]+",
          }}
          required
          fullWidth
        />
        <TextField
          id="displayName"
          label="Display name"
          margin="normal"
          value={this.state.object.displayName || ""}
          onChange={this.onChange}
          required
          fullWidth
        />
        <FormControl
          label="Gateways"
        >
          <FormGroup>
            <FormControlLabel
              label="Organization can have gateways"
              control={
                <Checkbox
                  id="canHaveGateways"
                  checked={!!this.state.object.canHaveGateways}
                  onChange={this.onChange}
                  value="true"
                  color="primary"
                />
              }
            />
          </FormGroup>
          <FormHelperText>When checked, it means that organization administrators are able to add their own gateways to the network. Note that the usage of the gateways is not limited to this organization.</FormHelperText>
        </FormControl>
      </Form>
    );
  }
}

export default OrganizationForm;
