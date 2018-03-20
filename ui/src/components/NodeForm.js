import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';

import Select from "react-select";

import Loaded from "./Loaded.js";
import DeviceProfileStore from "../stores/DeviceProfileStore";


class NodeForm extends Component {
  constructor() {
    super();

    this.state = {
      node: {},
      devEUIDisabled: false,
      disabled: false,
      deviceProfiles: [],
      loaded: {
        deviceProfiles: false,
      },
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentDidMount() {
    this.setState({
      node: this.props.node,
    });

    DeviceProfileStore.getAllForApplicationID(this.props.applicationID, 9999, 0, (totalCount, deviceProfiles) => {
      this.setState({
        deviceProfiles: deviceProfiles,
        loaded: {
          deviceProfiles: true,
        },
      });
    });
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      node: nextProps.node,
      devEUIDisabled: typeof nextProps.node.devEUI !== "undefined",
      disabled: nextProps.disabled,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.node);
  }

  onChange(field, e) {
    let node = this.state.node;
    if (e.target.type === "number") {
      node[field] = parseInt(e.target.value, 10); 
    } else if (e.target.type === "checkbox") {
      node[field] = e.target.checked;
    } else {
      node[field] = e.target.value;
    }
    this.setState({node: node});
  };

  onSelectChange(field, val) {
    let node = this.state.node;
    if (val !== null) {
      node[field] = val.value;
    } else {
      node[field] = null;
    }
    this.setState({
      node: node,
    });
  }

  render() {
    const deviceProfileOptions = this.state.deviceProfiles.map((deviceProfile, i) => {
      return {
        value: deviceProfile.deviceProfileID,
        label: deviceProfile.name,
      };
    });

    return (
      <Loaded loaded={this.state.loaded}>
        <form onSubmit={this.handleSubmit}>
          <div className={"alert alert-warning " + (this.state.deviceProfiles.length > 0 ? 'hidden' : '')}>
            No device-profiles are available, a <Link to={`/organizations/${this.props.organizationID}/device-profiles`}>device-profile</Link> needs to be created first for this organization.
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="name">Device name</label>
            <input className="form-control" id="name" type="text" placeholder="e.g. 'garden-sensor'" required value={this.state.node.name || ''} pattern="[\w-]+" onChange={this.onChange.bind(this, 'name')} />
            <p className="help-block">
              The name may only contain words, numbers and dashes.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="name">Device description</label>
            <input className="form-control" id="description" type="text" placeholder="a short description of your node" required value={this.state.node.description || ''} onChange={this.onChange.bind(this, 'description')} />
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="devEUI">Device EUI</label>
            <input className="form-control" id="devEUI" type="text" placeholder="0000000000000000" pattern="[A-Fa-f0-9]{16}" required disabled={this.state.devEUIDisabled} value={this.state.node.devEUI || ''} onChange={this.onChange.bind(this, 'devEUI')} /> 
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="deviceProfileID">Device-profile</label>
            <Select
              name="deviceProfileID"
              options={deviceProfileOptions}
              value={this.state.node.deviceProfileID}
              onChange={this.onSelectChange.bind(this, 'deviceProfileID')}
              required={true}
            />
          </div>
          <hr />
          <div className="btn-toolbar pull-right">
            <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
            <button type="submit" className={"btn btn-primary " + (this.state.disabled ? 'hidden' : '')}>Submit</button>
          </div>
        </form>
      </Loaded>
    );
  }
}

export default withRouter(NodeForm);
