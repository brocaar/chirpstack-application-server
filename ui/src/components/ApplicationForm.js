import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';
import Select from "react-select";
import {Controlled as CodeMirror} from "react-codemirror2";

import Loaded from "./Loaded.js";
import ServiceProfileStore from "../stores/ServiceProfileStore";
import "codemirror/mode/javascript/javascript";


class ApplicationForm extends Component {
  constructor() {
    super();
    this.state = {
      application: {},
      serviceProfiles: [],
      update: false,
      loaded: {
        serviceProfiles: false,
      },
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
        loaded: {
          serviceProfiles: true,
        },
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

  onCodeChange(field, editor, data, newCode) {
    let application = this.state.application;
    application[field] = newCode;
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

    const payloadCodecOptions = [
      {value: "", label: "None"},
      {value: "CAYENNE_LPP", label: "Cayenne LPP"},
      {value: "CUSTOM_JS", label: "Custom JavaScript codec functions"},
    ];

    const codeMirrorOptions = {
      lineNumbers: true,
      mode: "javascript",
      theme: 'base16-light',
    };
    
    let payloadEncoderScript = this.state.application.payloadEncoderScript;
    let payloadDecoderScript = this.state.application.payloadDecoderScript;

    if (payloadEncoderScript === "" || payloadEncoderScript === undefined) {
      payloadEncoderScript = `// Encode encodes the given object into an array of bytes.
//  - fPort contains the LoRaWAN fPort number
//  - obj is an object, e.g. {"temperature": 22.5}
// The function must return an array of bytes, e.g. [225, 230, 255, 0]
function Encode(fPort, obj) {
  return [];
}`;
    }

    if (payloadDecoderScript === "" || payloadDecoderScript === undefined) {
      payloadDecoderScript = `// Decode decodes an array of bytes into an object.
//  - fPort contains the LoRaWAN fPort number
//  - bytes is an array of bytes, e.g. [225, 230, 255, 0]
// The function must return an object, e.g. {"temperature": 22.5}
function Decode(fPort, bytes) {
  return {};
}`;
    }

    return (
      <Loaded loaded={this.state.loaded}>
        <form onSubmit={this.handleSubmit}>
          <div className={"alert alert-warning " + (this.state.serviceProfiles.length > 0 ? 'hidden' : '')}>
            No service-profiles are associated with this organization, a <Link to={`/organizations/${this.props.organizationID}/service-profiles`}>service-profile</Link> needs to be created first for this organization.
          </div>
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
              required={true}
            />
            <p className="help-block">
              The service-profile to which this application will be attached. Note that you can't change this value after the application has been created.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="payloadCodec">Payload codec</label>
            <Select
              name="payloadCodec"
              options={payloadCodecOptions}
              value={this.state.application.payloadCodec}
              onChange={this.onSelectChange.bind(this, 'payloadCodec')}
            />
            <p className="help-block">
              By defining a payload codec, LoRa App Server can encode and decode the binary device payload for you.
            </p>
          </div>
          <div className={"form-group " + (this.state.application.payloadCodec === "CUSTOM_JS" ? "" : "hidden")}>
            <label className="control-label" htmlFor="payloadDecoderScript">Payload decoder function</label>
            <CodeMirror
              value={payloadDecoderScript}
              options={codeMirrorOptions}
              onBeforeChange={this.onCodeChange.bind(this, 'payloadDecoderScript')}
            />
            <p className="help-block">
              The function must have the signature <strong>function Decode(fPort, bytes)</strong> and must return an object.
              LoRa App Server will convert this object to JSON.
            </p>
          </div>
          <div className={"form-group " + (this.state.application.payloadCodec === "CUSTOM_JS" ? "" : "hidden")}>
            <label className="control-label" htmlFor="payloadEncoderScript">Payload encoder function</label>
            <CodeMirror
              value={payloadEncoderScript}
              options={codeMirrorOptions}
              onBeforeChange={this.onCodeChange.bind(this, 'payloadEncoderScript')}
            />
            <p className="help-block">
              The function must have the signature <strong>function Encode(fPort, obj)</strong> and must return an array
              of bytes.
            </p>
          </div>
          <hr />
          <div className="btn-toolbar pull-right">
            <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
            <button type="submit" className="btn btn-primary">Submit</button>
          </div>
        </form>
      </Loaded>
    );
  }
}

export default withRouter(ApplicationForm);
