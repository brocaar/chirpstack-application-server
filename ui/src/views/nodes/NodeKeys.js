import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import NodeStore from "../../stores/NodeStore";


class DeviceKeysForm extends Component {
  constructor() {
    super();

    this.state = {
      deviceKeys: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentDidMount() {
    this.setState({
      deviceKeys: this.props.deviceKeys,
    });
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      deviceKeys: nextProps.deviceKeys,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.deviceKeys);
  }

  onChange(field, e) {
    let deviceKeys = this.state.deviceKeys;
    deviceKeys[field] = e.target.value;

    this.setState({
      deviceKeys: deviceKeys,
    });
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="nwkKey">Network key</label>
          <input className="form-control" id="nwkKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.deviceKeys.nwkKey || ''} onChange={this.onChange.bind(this, 'nwkKey')} /> 
          <p className="help-block">
            For LoRaWAN 1.0 devices, this is the only key you need to set (in LoRaWAN 1.0 this used to be the application-key).
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appKey">Application key</label>
          <input className="form-control" id="appKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" value={this.state.deviceKeys.appKey || ''} onChange={this.onChange.bind(this, 'appKey')} /> 
          <p className="help-block">
            Leave this blank for LoRaWAN 1.0 devices.
          </p>
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
          <button type="submit" className={"btn btn-primary " + (this.state.disabled ? 'hidden' : '')}>Submit</button>
        </div>
      </form>
    );
  }
}

class NodeKeys extends Component {
  constructor() {
    super();

    this.state = {
      deviceKeys: {},
      update: false,
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  componentDidMount() {
    NodeStore.getNodeKeys(this.props.match.params.devEUI, (deviceKeys) => {
      this.setState({
        update: true,
        deviceKeys: deviceKeys.deviceKeys,
      });
    });
  }

  onSubmit(deviceKeys) {
    if (this.state.update) {
      NodeStore.updateNodeKeys(this.props.match.params.devEUI, {deviceKeys: deviceKeys}, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
      });
    } else {
      NodeStore.createNodeKeys(this.props.match.params.devEUI, {deviceKeys: deviceKeys}, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
      });
    }
  }

  render() {
    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-body">
            <DeviceKeysForm history={this.props.history} deviceKeys={this.state.deviceKeys} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(NodeKeys);
