import React, { Component } from 'react';

import Select from "react-select";

import ServiceProfileStore from "../stores/ServiceProfileStore";


class ApplicationForm extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.state = {
      application: {},
      serviceProfiles: [],
      update: false,
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentDidMount() {
    this.setState({
      application: this.props.application,
    });

    ServiceProfileStore.getAllForOrganizationID(this.props.organizationID, 9999, 0, (totalCount, serviceProfiles) => {
      this.setState({
        serviceProfiles: serviceProfiles,
      });
    });
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      application: nextProps.application,
      update: nextProps.application.id !== undefined,
    });
  }

  onChange(field, e) {
    let application = this.state.application;
    if (e.target.type === "number") {
      application[field] = parseInt(e.target.value, 10); 
    } else if (e.target.type === "checkbox") {
      application[field] = e.target.checked;
    } else {
      application[field] = e.target.value;
    }
    this.setState({application: application});
  }

  onSelectChange(field, val) {
    let application = this.state.application;
    if (val !== null) {
      application[field] = val.value;
    } else {
      application[field] = null;
    }
    this.setState({
      application: application,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.application);
  }

  render() {
    const serviceProfileOptions = this.state.serviceProfiles.map((serviceProfile, i) => {
      return {
        value: serviceProfile.serviceProfileID,
        label: serviceProfile.name,
      };
    });

    return (
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Application name</label>
          <input className="form-control" id="name" type="text" placeholder="e.g. 'temperature-sensor'" pattern="[\w-]+" required value={this.state.application.name || ''} onChange={this.onChange.bind(this, 'name')} />
          <p className="help-block">
            The name may only contain words, numbers and dashes. 
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Application description</label>
          <input className="form-control" id="description" type="text" placeholder="a short description of your application" required value={this.state.application.description || ''} onChange={this.onChange.bind(this, 'description')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="serviceProfileID">Service-profile</label>
          <Select
            name="serviceProfileID"
            options={serviceProfileOptions}
            value={this.state.application.serviceProfileID}
            onChange={this.onSelectChange.bind(this, 'serviceProfileID')}
            disabled={this.state.update}
          />
          <p className="help-block">
            The service-profile to which this application will be attached. Note that you can't change this value after the application has been created.
          </p>
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.context.router.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

export default ApplicationForm;
