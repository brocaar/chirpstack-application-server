import React from "react";

import Typography from '@material-ui/core/Typography';
import TextField from '@material-ui/core/TextField';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormControl from '@material-ui/core/FormControl';
import FormHelperText from '@material-ui/core/FormHelperText';
import Checkbox from '@material-ui/core/Checkbox';

import FormComponent from "../../classes/FormComponent";
import Form from "../../components/Form";


class OrganizationUserForm extends FormComponent {
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
            label="Email"
            id="email"
            margin="normal"
            value={this.state.object.email || ""}
            onChange={this.onChange}
            required
            fullWidth
            disabled={this.props.update}
          />
          <Typography variant="body1">
            An user without additional permissions will be able to see all
            resources under this organization and will be able to send and
            receive device payloads.
          </Typography>
          <FormControl fullWidth margin="normal">
            <FormControlLabel
              label="User is organization admin"
              control={
                <Checkbox
                  id="isAdmin"
                  checked={!!this.state.object.isAdmin}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
            <FormHelperText>An organization admin user is able to add and modify resources part of the organization.</FormHelperText>
          </FormControl>
          {!!!this.state.object.isAdmin && <FormControl fullWidth margin="normal">
            <FormControlLabel
              label="User is device admin"
              control={
                <Checkbox
                  id="isDeviceAdmin"
                  checked={!!this.state.object.isDeviceAdmin}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
            <FormHelperText>A device admin user is able to add and modify resources part of the organization that are related to devices.</FormHelperText>
          </FormControl>}
          {!!!this.state.object.isAdmin && <FormControl fullWidth margin="normal">
            <FormControlLabel
              label="User is gateway admin"
              control={
                <Checkbox
                  id="isGatewayAdmin"
                  checked={!!this.state.object.isGatewayAdmin}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
            <FormHelperText>A gateway admin user is able to add and modify gateways part of the organization.</FormHelperText>
          </FormControl>}
      </Form>
    );
  }
}

export default OrganizationUserForm;
