import React, { Component } from "react";
import { withRouter } from 'react-router-dom';

import NodeStore from "../../stores/NodeStore";


class NodeActivationForm extends Component {
  constructor() {
    super();

    this.state = {
      activation: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.getRandomDevAddr = this.getRandomDevAddr.bind(this);
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.activation);
  }

  onChange(field, e) {
    let activation = this.state.activation;
    if (e.target.type === "number") {
      activation[field] = parseInt(e.target.value, 10); 
    } else if (e.target.type === "checkbox") {
      activation[field] = e.target.checked;
    } else {
      activation[field] = e.target.value;
    }
    this.setState({activation: activation});
  }

  getRandomDevAddr(e) {
    e.preventDefault();

    NodeStore.getRandomDevAddr(this.props.devEUI, (responseData) => {
      let activation = this.state.activation;
      activation["devAddr"] = responseData.devAddr;
      this.setState({
        activation: activation,
      });
    });
  }

  getRandomKey(field, e) {
    e.preventDefault();

    let activation = this.state.activation;
    let key = "";
    const possible = 'abcdef0123456789';

    for(let i = 0; i < 32; i++){
      key += possible.charAt(Math.floor(Math.random() * possible.length));
    }

    activation[field] = key;

    this.setState({
      activation: activation,
    });
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="devAddr">Device address</label> (<a href="" onClick={this.getRandomDevAddr}>generate</a>)
          <input className="form-control" id="devAddr" type="text" placeholder="00000000" pattern="[a-fA-F0-9]{8}" required value={this.state.activation.devAddr || ''} onChange={this.onChange.bind(this, 'devAddr')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="nwkSEncKey">Network session encryption key</label> (<a href="" onClick={this.getRandomKey.bind(this, "nwkSEncKey")}>generate</a>)
          <input className="form-control" id="nwkSEncKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.nwkSEncKey || ''} onChange={this.onChange.bind(this, 'nwkSEncKey')} />
          <p className="help-block">
            For LoRaWAN 1.0 devices, set this to the network session key.
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="sNwkSIntKey">Serving network session integrity key</label> (<a href="" onClick={this.getRandomKey.bind(this, "sNwkSIntKey")}>generate</a>)
          <input className="form-control" id="sNwkSIntKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.sNwkSIntKey || ''} onChange={this.onChange.bind(this, 'sNwkSIntKey')} />
          <p className="help-block">
            For LoRaWAN 1.0 devices, set this to the network session key.
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="fNwkSIntKey">Forwarding network session integrity key</label> (<a href="" onClick={this.getRandomKey.bind(this, "fNwkSIntKey")}>generate</a>)
          <input className="form-control" id="fNwkSIntKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.fNwkSIntKey || ''} onChange={this.onChange.bind(this, 'fNwkSIntKey')} />
          <p className="help-block">
            For LoRaWAN 1.0 devices, set this to the network session key.
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appSKey">Application session key</label> (<a href="" onClick={this.getRandomKey.bind(this, "appSKey")}>generate</a>)
          <input className="form-control" id="appSKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.appSKey || ''}  onChange={this.onChange.bind(this, 'appSKey')} />
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

class ActivateNode extends Component {
  constructor() {
    super();
    this.state = {
      activation: {},
      node: {},
    };

    this.onSubmit = this.onSubmit.bind(this);
  }

  onSubmit(activation) {
    NodeStore.activateNode(this.props.match.params.devEUI, activation, (responseData) => {
      this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`);
    });
  }

  render() {
    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-body">
            <NodeActivationForm history={this.props.history} devEUI={this.props.match.params.devEUI} onSubmit={this.onSubmit} />
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(ActivateNode);
