import React, { Component } from "react";

import TextField from "@material-ui/core/TextField";

class DurationField extends Component {
  constructor() {
    super();

    this.state = {
      value: 0,
    };
  }


  onChange = (e) => {
    this.setState({
      value: e.target.value,
    });

    this.props.onChange({
      target: {
        value: `${e.target.value}s`,
        type: "text",
        id: this.props.id,
      },
    });
  }

  componentDidMount() {
    const str = this.props.value || "";
    this.setState({
      value: str.replace(/s/, ''),
    });
  }

  render() {
    return(
      <TextField
        type="number"
        id={this.props.id}
        label={this.props.label}
        value={this.state.value}
        helperText={this.props.helperText}
        margin="normal"
        onChange={this.onChange}
        required
        fullWidth
      />
    );
  }
}

export default DurationField;

