import React from "react";

import TextField from '@material-ui/core/TextField';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormGroup from "@material-ui/core/FormGroup";
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
        <FormGroup>
          <TextField
            label="Username"
            margin="normal"
            value={this.state.object.username || ""}
            required
            fullWidth
            disabled
          />
          <FormControlLabel
            label="Is organization admin"
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
      </Form>
    );
  }
}

export default OrganizationUserForm;
