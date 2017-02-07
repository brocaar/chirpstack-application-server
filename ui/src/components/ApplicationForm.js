import React, { Component } from 'react';

class ApplicationForm extends Component {
  constructor() {
    super();
    this.state = {
      application: {},
      nameDisabled: false,
    };
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      application: nextProps.application,
      nameDisabled: typeof nextProps.application.name !== "undefined",
    });
  }

  onChange(field, e) {
    let application = this.state.application;
    application[field] = e.target.value;
    this.setState({application: application});
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.application);
  }

  render() {
    return (
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Application name</label>
          <input className="form-control" id="name" type="text" placeholder="e.g. 'temperature-sensor'" pattern="[\w-]+" required value={this.state.application.name || ''} disabled={this.state.nameDisabled} onChange={this.onChange.bind(this, 'name')} />
          <p className="help-block">
            The name may only contain words, numbers and dashes. This name will be used as identifier and will be used for MQTT topics.
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Application description</label>
          <input className="form-control" id="description" type="text" placeholder="a short description of your application" required value={this.state.application.description || ''} onChange={this.onChange.bind(this, 'description')} />
        </div>
        <hr />
        <button type="submit" className="btn btn-primary pull-right">Submit</button>
      </form>
    );
  }
}

export default ApplicationForm;
