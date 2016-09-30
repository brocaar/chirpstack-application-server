import React, { Component } from 'react';

import ChannelStore from "../stores/ChannelStore";

class NodeForm extends Component {
  constructor() {
    super();

    this.state = {node: {}, devEUIDisabled: false, channelLists: []};
    this.handleSubmit = this.handleSubmit.bind(this);

    ChannelStore.getAllChannelLists((lists) => {
      this.setState({channelLists: lists});
    });
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      node: nextProps.node,
      devEUIDisabled: typeof nextProps.node.devEUI !== "undefined",
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
    } else {
      node[field] = e.target.value;
    }
    this.setState({node: node});
  };
  
  render() {
    return (
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Node name</label>
          <input className="form-control" id="name" type="text" placeholder="a descriptive name for your node" required value={this.state.node.name || ''} onChange={this.onChange.bind(this, 'name')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="devEUI">Device EUI</label>
          <input className="form-control" id="devEUI" type="text" placeholder="0000000000000000" pattern="[A-Fa-f0-9]{16}" required disabled={this.state.devEUIDisabled} value={this.state.node.devEUI || ''} onChange={this.onChange.bind(this, 'devEUI')} /> 
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appEUI">Application EUI</label>
          <input className="form-control" id="appEUI" type="text" placeholder="0000000000000000" pattern="[A-Fa-f0-9]{16}" required value={this.state.node.appEUI || ''} onChange={this.onChange.bind(this, 'appEUI')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appKey">Application key</label>
          <input className="form-control" id="appKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required value={this.state.node.appKey || ''} onChange={this.onChange.bind(this, 'appKey')} />
        </div>
        <div className="form-group">
          <label className="control-label">Receive window</label>
          <div className="radio">
            <label>
              <input type="radio" name="rxWindow" id="rxWindow1" value="RX1" checked={this.state.node.rxWindow === "RX1"} onChange={this.onChange.bind(this, 'rxWindow')} /> RX1
            </label>
          </div>
          <div className="radio">
            <label>
              <input type="radio" name="rxWindow" id="rxWindow2" value="RX2" checked={this.state.node.rxWindow === "RX2"} onChange={this.onChange.bind(this, 'rxWindow')} /> RX2 (one second after RX1)
            </label>
          </div>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rxDelay">Receive window delay</label>
          <input className="form-control" id="rxDelay" type="number" min="0" max="15" required value={this.state.node.rxDelay || 0} onChange={this.onChange.bind(this, 'rxDelay')} />
          <p className="help-block">The delay in seconds (0-15) between the end of the TX uplink and the opening of the first reception slot (0=1 sec, 1=1 sec, 2=2 sec, 3=3 sec, ... 15=15 sec).</p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx1DROffset">RX1 data-rate offset</label>
          <input className="form-control" id="rx1DROffset" type="number" required value={this.state.node.rx1DROffset || 0} onChange={this.onChange.bind(this, 'rx1DROffset')} />
          <p className="help-block">
            Sets the offset between the uplink data rate and the downlink data-rate used to communicate with the end-device on the first reception slot (RX1).
            Please refer to the LoRaWAN specs for the values that are valid in your region.
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx2DR">RX2 data-rate</label>
          <input className="form-control" id="rx2DR" type="number" required value={this.state.node.rx2DR || 0} onChange={this.onChange.bind(this, 'rx2DR')} />
          <p className="help-block">
            The data-rate to use when RX2 is used as receive window.
            Please refer to the LoRaWAN specs for the values that are valid in your region.
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="channelListID">Channel-list</label>
          <select className="form-control" id="channelListID" name="channelListID" value={this.state.node.channelListID} onChange={this.onChange.bind(this, "channelListID")}>
            <option value="0"></option>
            {
              this.state.channelLists.map((cl, i) => {
                return (<option key={cl.id} value={cl.id}>{cl.name}</option>);
              })
            }
          </select>
          <p className="help-block">
            Some LoRaWAN ISM bands implement an optional channel-frequency list that can be sent when using OTAA.
            Please refer to the LoRaWAN specs for the values that are valid in your region.
          </p>
        </div>
        <hr />
        <button type="submit" className="btn btn-primary pull-right">Submit</button>
      </form>
    );
  }
}

export default NodeForm;
