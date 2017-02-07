import React, { Component } from "react";
import ErrorStore from "../stores/ErrorStore";
import dispatcher from "../dispatcher";


class ErrorLine extends Component {
  constructor() {
    super();
    this.handleDelete = this.handleDelete.bind(this);
  }

  handleDelete() {
    dispatcher.dispatch({
      type: "DELETE_ERROR",
      id: this.props.id,
    });
  }

  render() {
    return (
      <div className="alert alert-danger">
        <button type="button" className="close" onClick={this.handleDelete}><span>&times;</span></button>
        <strong>Error</strong> {this.props.error.error} (code: {this.props.error.code})
      </div>
    )
  }
}


class Errors extends Component {
  constructor() {
    super();
    this.state = {
      errors: ErrorStore.getAll(),
    };
  }

  componentWillMount() {
    ErrorStore.on("change", () => {
      this.setState({
        errors: ErrorStore.getAll(),
      });
    });
  }

  render() {
    const ErrorLines = this.state.errors.map((error, i) => <ErrorLine key={error.id} id={error.id} error={error.error} />);

    return (
          <div>
            {ErrorLines}
          </div>
        )
  }
}

export default Errors;
