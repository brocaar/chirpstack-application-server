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
    this.getRandomAppSKey = this.getRandomAppSKey.bind(this);
    this.getRandomNwkSKey = this.getRandomNwkSKey.bind(this);
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

  getRandomNwkSKey(e) {
    e.preventDefault();

    let nwkSKey = '';
    const possible = 'abcdef0123456789';
    for(let i = 0; i < 32; i++){
      nwkSKey += possible.charAt(Math.floor(Math.random() * possible.length));
    }
    let activation = this.state.activation;
    activation["nwkSKey"] = nwkSKey;
    this.setState({activation: activation});
  }

  getRandomAppSKey(e) {
    e.preventDefault();

    let appSKey = '';
    const possible = 'abcdef0123456789';
    for(let i = 0; i < 32; i++){
      appSKey += possible.charAt(Math.floor(Math.random() * possible.length));
    }
    let activation = this.state.activation;
    activation["appSKey"] = appSKey;
    this.setState({activation: activation});
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="devAddr">Device address</label> (<a href="" onClick={this.getRandomDevAddr}>generate</a>)
          <input className="form-control" id="devAddr" type="text" placeholder="00000000" pattern="[a-fA-F0-9]{8}" required value={this.state.activation.devAddr || ''} onChange={this.onChange.bind(this, 'devAddr')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="nwkSKey">Network session key</label> (<a href="" onClick={this.getRandomNwkSKey}>generate</a>)
          <input className="form-control" id="nwkSKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.nwkSKey || ''} onChange={this.onChange.bind(this, 'nwkSKey')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appSKey">Application session key</label> (<a href="" onClick={this.getRandomAppSKey}>generate</a>)
          <input className="form-control" id="appSKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.appSKey || ''}  onChange={this.onChange.bind(this, 'appSKey')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx2DR">Uplink frame-counter</label>
          <input className="form-control" id="fCntUp" type="number" min="0" required value={this.state.activation.fCntUp || 0} onChange={this.onChange.bind(this, 'fCntUp')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx2DR">Downlink frame-counter</label>
          <input className="form-control" id="fCntDown" type="number" min="0" required value={this.state.activation.fCntDown || 0} onChange={this.onChange.bind(this, 'fCntDown')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="skipFCntCheck">Disable frame-counter validation</label>
          <div className="checkbox">
            <label>
              <input type="checkbox" name="skipFCntCheck" id="skipFCntCheck" checked={!!this.state.activation.skipFCntCheck} onChange={this.onChange.bind(this, 'skipFCntCheck')} /> Disable frame-counter validation
            </label>
          </div>
          <p className="help-block">
            Note that disabling the frame-counter validation will compromise security as it enables people to perform replay-attacks.
          </p>
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
