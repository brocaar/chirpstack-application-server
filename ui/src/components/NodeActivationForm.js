import React, { Component } from 'react';

import NodeStore from "../stores/NodeStore";

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
    } else {
      activation[field] = e.target.value;
    }
    this.setState({activation: activation});
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      activation: nextProps.activation,
    });
  }

  getRandomDevAddr(e) {
    e.preventDefault();

    if (!this.props.node.isABP || this.props.disabled) {
      return; 
    }

    NodeStore.getRandomDevAddr((responseData) => {
      let activation = this.state.activation;
      activation["devAddr"] = responseData.devAddr;
      this.setState({activation: activation});
    });
  }

  render() {
    return(
      <div>
        <form onSubmit={this.handleSubmit}>
          <fieldset disabled={this.props.disabled}>
            <div className="form-group">
              <label className="control-label" htmlFor="devAddr">Device address</label> (<a href="" onClick={this.getRandomDevAddr}>generate</a>)
              <input className="form-control" id="devAddr" type="text" placeholder="00000000" pattern="[a-fA-F0-9]{8}" required disabled={!this.props.node.isABP} value={this.state.activation.devAddr || ''} onChange={this.onChange.bind(this, 'devAddr')} />
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="nwkSKey">Network session key</label>
              <input className="form-control" id="nwkSKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.nwkSKey || ''} disabled={!this.props.node.isABP} onChange={this.onChange.bind(this, 'nwkSKey')} />
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="appSKey">Application session key</label>
              <input className="form-control" id="appSKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.activation.appSKey || ''} disabled={!this.props.node.isABP} onChange={this.onChange.bind(this, 'appSKey')} />
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rx2DR">Uplink frame-counter</label>
              <input className="form-control" id="fCntUp" type="number" min="0" required value={this.state.activation.fCntUp || 0} disabled={!this.props.node.isABP} onChange={this.onChange.bind(this, 'fCntUp')} />
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rx2DR">Downlink frame-counter</label>
              <input className="form-control" id="fCntDown" type="number" min="0" required value={this.state.activation.fCntDown || 0} disabled={!this.props.node.isABP} onChange={this.onChange.bind(this, 'fCntDown')} />
            </div>
            <hr />
            <button type="submit" className={"btn btn-primary pull-right " + ((!this.props.node.isABP || this.props.disabled) ? 'hidden' : '')}>Submit</button>
          </fieldset>
        </form>
      </div>
    );
  }
}

export default NodeActivationForm;
