import React, { Component } from 'react';
import NodeSessionStore from "../stores/NodeSessionStore";

class NodeSessionForm extends Component {
  constructor() {
    super();

    this.state = {session: {}, devEUIDisabled: false};
    this.handleSubmit = this.handleSubmit.bind(this);
    this.getRandomDevAddr = this.getRandomDevAddr.bind(this);
  }

  onChange(field, e) {
    let session = this.state.session;
    if (e.target.type === "number") {
      session[field] = parseInt(e.target.value, 10); 
    } else {
      session[field] = e.target.value;
    }
    this.setState({session: session});
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      session: nextProps.session,
      devAddrDisabled: typeof nextProps.session.devAddr !== "undefined",
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.session);
  }

  getRandomDevAddr(e) {
    e.preventDefault();

    if (this.state.devAddrDisabled) {
      return;
    }

    NodeSessionStore.getRandomDevAddr((responseData) => {
      let session = this.state.session;
      session["devAddr"] = responseData.devAddr;
      this.setState({session: session});
    });
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="devAddr">DevAddr</label> (<a href="" onClick={this.getRandomDevAddr}>generate</a>)
          <input className="form-control" id="devAddr" type="text" required disabled={this.state.devAddrDisabled} value={this.state.session.devAddr || ''} onChange={this.onChange.bind(this, 'devAddr')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="nwkSKey">NwkSKey</label>
          <input className="form-control" id="nwkSKey" type="text" required value={this.state.session.nwkSKey || ''} onChange={this.onChange.bind(this, 'nwkSKey')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="appSKey">AppSKey</label>
          <input className="form-control" id="appSKey" type="text" required value={this.state.session.appSKey || ''} onChange={this.onChange.bind(this, 'appSKey')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx2DR">FCnt up</label>
          <input className="form-control" id="fCntUp" type="number" required value={this.state.session.fCntUp || 0} onChange={this.onChange.bind(this, 'fCntUp')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="rx2DR">FCnt down</label>
          <input className="form-control" id="fCntDown" type="number" required value={this.state.session.fCntDown || 0} onChange={this.onChange.bind(this, 'fCntDown')} />
        </div>
        <hr />
        <button type="submit" className="btn btn-primary pull-right">Submit</button>
      </form>
    );
  }
}

export default NodeSessionForm;
