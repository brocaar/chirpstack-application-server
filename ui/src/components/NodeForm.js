import React, { Component } from 'react';

class NodeForm extends Component {
  constructor() {
    super();

    this.state = {node: {}, devEUIDisabled: false};
    this.handleSubmit = this.handleSubmit.bind(this);
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
          <input className="form-control" id="name" type="text" required value={this.state.node.name || ''} onChange={this.onChange.bind(this, 'name')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="devEUI">DevEUI</label>
          <input className="form-control" id="devEUI" type="text" required disabled={this.state.devEUIDisabled} value={this.state.node.devEUI || ''} onChange={this.onChange.bind(this, 'devEUI')} /> 
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appEUI">AppEUI</label>
          <input className="form-control" id="appEUI" type="text" required value={this.state.node.appEUI || ''} onChange={this.onChange.bind(this, 'appEUI')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appKey">AppKey</label>
          <input className="form-control" id="appKey" type="text" required value={this.state.node.appKey || ''} onChange={this.onChange.bind(this, 'appKey')} />
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
              <input type="radio" name="rxWindow" id="rxWindow2" value="RX2" checked={this.state.node.rxWindow === "RX2"} onChange={this.onChange.bind(this, 'rxWindow')} /> RX2
            </label>
          </div>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rxDelay">RXDelay</label>
          <input className="form-control" id="rxDelay" type="number" required value={this.state.node.rxDelay || 0} onChange={this.onChange.bind(this, 'rxDelay')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx1DROffset">RX1DRoffset</label>
          <input className="form-control" id="rx1DROffset" type="number" required value={this.state.node.rx1DROffset || 0} onChange={this.onChange.bind(this, 'rx1DROffset')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx2DR">RX2DR</label>
          <input className="form-control" id="rx2DR" type="number" required value={this.state.node.rx2DR || 0} onChange={this.onChange.bind(this, 'rx2DR')} />
        </div>
        <hr />
        <div className="form-group">
          <button type="submit" className="btn btn-primary pull-right">Submit</button>
        </div>
      </form>
    );
  }
}

export default NodeForm;
