import React, { Component } from "react";

import NodeStore from "../../stores/NodeStore";


class NodeActivation extends Component {
  constructor() {
    super();

    this.state = {
      activation: {},
    };
  }

  componentDidMount() {
    NodeStore.getActivation(this.props.match.params.devEUI, (nodeActivation) => {
      this.setState({
        activation: nodeActivation,
      });
    });
  }

  render() {
    if (this.state.activation.devAddr === undefined) {
      return(
        <div className="panel panel-default">
          <div className="panel-body">
            <div>
              The node has not been activated yet or device has been inactive for a long time.
            </div>
          </div>
        </div>
      );
    } else {
      return(
        <div className="panel panel-default">
          <div className="panel-body">
            <form onSubmit={this.handleSubmit}>
              <fieldset disabled={true}>
                <div className="form-group">
                  <label className="control-label" htmlFor="devAddr">Device address</label>
                  <input className="form-control" id="devAddr" type="text" value={this.state.activation.devAddr || ''} />
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="nwkSEncKey">Network session encryption key</label>
                  <input className="form-control" id="nwkSEncKey" type="text" value={this.state.activation.nwkSEncKey || ''} />
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="sNwkSIntKey">Serving network session integrity key</label>
                  <input className="form-control" id="sNwkSIntKey" type="text" value={this.state.activation.sNwkSIntKey || ''} />
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="fNwkSIntKey">Forwarding network session integrity key</label>
                  <input className="form-control" id="fNwkSIntKey" type="text" value={this.state.activation.fNwkSIntKey || ''} />
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="appSKey">Application session key</label>
                  <input className="form-control" id="appSKey" type="text" value={this.state.activation.appSKey || ''} />
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="rx2DR">Uplink frame-counter</label>
                  <input className="form-control" id="fCntUp" type="number" value={this.state.activation.fCntUp || 0} />
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="rx2DR">Downlink frame-counter</label>
                  <input className="form-control" id="fCntDown" type="number" required value={this.state.activation.fCntDown || 0} />
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="skipFCntCheck">Disable frame-counter validation</label>
                  <div className="checkbox">
                    <label>
                      <input type="checkbox" name="skipFCntCheck" id="skipFCntCheck" checked={!!this.state.activation.skipFCntCheck} /> Disable frame-counter validation
                    </label>
                  </div>
                  <p className="help-block">
                    Note that disabling the frame-counter validation will compromise security as it enables people to perform replay-attacks.
                    This setting can only be set for ABP devices.
                  </p>
                </div>
              </fieldset>
            </form>
          </div>
        </div>
      );
    }
  }
}

export default NodeActivation;