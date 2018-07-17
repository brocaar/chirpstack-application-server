import React from "react";

import TextField from '@material-ui/core/TextField';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormGroup from "@material-ui/core/FormGroup";
import Checkbox from '@material-ui/core/Checkbox';

import FormComponent from "../../classes/FormComponent";
import FormControl from "../../components/FormControl";
import Form from "../../components/Form";


class UserForm extends FormComponent {
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
          id="username"
          label="Username"
          margin="normal"
          value={this.state.object.username || ""}
          onChange={this.onChange}
          required
          fullWidth
        />
        <TextField
          id="email"
          label="E-mail address"
          margin="normal"
          value={this.state.object.email || ""}
          onChange={this.onChange}
          required
          fullWidth
        />
        <TextField
          id="note"
          label="Optional note"
          helperText="Optional note, e.g. a phone number, address, comment..."
          margin="normal"
          value={this.state.object.note || ""}
          onChange={this.onChange}
          rows={4}
          fullWidth
          multiline
        />
        {this.state.object.id === undefined && <TextField
          id="password"
          label="Password"
          type="password"
          margin="normal"
          value={this.state.object.password || ""}
          onChange={this.onChange}
          required
          fullWidth
        />}
        <FormControl label="Permissions">
          <FormGroup>
            <FormControlLabel
              label="Is active"
              control={
                <Checkbox
                  id="isActive"
                  checked={!!this.state.object.isActive}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
            <FormControlLabel
              label="Is global admin"
              control={
                <Checkbox
                  id="isAdmin"
                  checked={!!this.state.object.isAdmin}
                  onChange={this.onChange}
                  color="primary"
                />
              }
            />
          </FormGroup>
        </FormControl>
      </Form>
    );
  }
}

export default UserForm;
