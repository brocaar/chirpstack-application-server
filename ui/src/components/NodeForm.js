import React, { Component } from 'react';

import ChannelStore from "../stores/ChannelStore";

class NodeForm extends Component {
  constructor() {
    super();

    this.state = {
      activeTab: "node",
      node: {},
      devEUIDisabled: false,
      channelLists: [],
      disabled: false,
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.changeTab = this.changeTab.bind(this);
    this.updateNodeSettings = this.updateNodeSettings.bind(this);

    ChannelStore.getAllChannelLists((lists) => {
      this.setState({channelLists: lists});
    });
  }

  componentWillReceiveProps(nextProps) {
    let node = this.updateNodeSettings(nextProps.application, nextProps.node);

    this.setState({
      node: node,
      application: nextProps.application,
      devEUIDisabled: typeof nextProps.node.devEUI !== "undefined",
      disabled: nextProps.disabled,
    });

  }

  handleSubmit(e) {
    e.preventDefault();
    let node = this.state.node;

    if(node.isABP) {
      node.appKey = "00000000000000000000000000000000";
    }
    this.props.onSubmit(node);
  }

  changeTab(e) {
    e.preventDefault();
    this.setState({
      activeTab: e.target.getAttribute('aria-controls'),
    });
  }

  updateNodeSettings(application, node) {
    if (node.useApplicationSettings === true) {
      node.rxDelay = application.rxDelay;
      node.rx1DROffset = application.rx1DROffset;
      node.channelListID = application.channelListID;
      node.rxWindow = application.rxWindow;
      node.rx2DR = application.rx2DR;
      node.relaxFCnt = application.relaxFCnt;
      node.adrInterval = application.adrInterval;
      node.installationMargin = application.installationMargin;
      node.isABP = application.isABP;
      node.isClassC = application.isClassC;
    }
    return node;
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
    node = this.updateNodeSettings(this.state.application, node);
    this.setState({node: node});
  };

  render() {
    return (
      <div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={(this.state.activeTab === "node" ? 'active' : '')}><a onClick={this.changeTab} href="#node" aria-controls="node">Node details</a></li>
          <li role="presentation" className={(this.state.activeTab === "advanced-network-settings" ? 'active' : '')}><a onClick={this.changeTab} href="#advanced-network-settings" aria-controls="advanced-network-settings">Advanced network settings</a></li>
        </ul>
        <hr />
        <form onSubmit={this.handleSubmit}>
          <fieldset disabled={this.state.disabled}>
            <div className={(this.state.activeTab === "node" ? '' : 'hidden')}>
            <div className="form-group">
              <label className="control-label" htmlFor="name">Node name</label>
              <input className="form-control" id="name" type="text" placeholder="e.g. 'garden-sensor'" required value={this.state.node.name || ''} pattern="[\w-]+" onChange={this.onChange.bind(this, 'name')} />
              <p className="help-block">
                The name may only contain words, numbers and dashes.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="name">Node description</label>
              <input className="form-control" id="description" type="text" placeholder="a short description of your node" required value={this.state.node.description || ''} onChange={this.onChange.bind(this, 'description')} />
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
              <label className="control-label">Use application settings</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="useApplicationSettings" id="useApplicationSettings" checked={this.state.node.useApplicationSettings} onChange={this.onChange.bind(this, 'useApplicationSettings')} /> Use application settings
                </label>
              </div>
              <p className="help-block">
                When checked, it means that the node will use the (network) settings as set by the application.
                In case this node requires node-specific (network) settings, uncheck this box.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label">Class-C node</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="isClassC" id="isClassC" disabled={this.state.node.useApplicationSettings} checked={this.state.node.isClassC} onChange={this.onChange.bind(this, 'isClassC')} /> Class-C node
                </label>
              </div>
              <p className="help-block">
                When checked, it means that the node operates in Class-C mode (always listening) and that data will be sent directly to the node. <br/>
                In any other case, the data will be sent as soon as a receive window occurs.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label">ABP (activation by personalisation)</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="isABP" id="isABP" disabled={this.state.node.useApplicationSettings} checked={this.state.node.isABP} onChange={this.onChange.bind(this, 'isABP')} /> ABP activation
                </label>
              </div>
              <p className="help-block">When checked, it means that the node will be manually activated and that over-the-air activation (OTAA) will be disabled.</p>
            </div>
            <div className={"form-group " + (this.state.node.isABP ? 'hidden': '')}>
              <label className="control-label" htmlFor="appKey">Application key</label>
              <input className="form-control" id="appKey" type="text" placeholder="00000000000000000000000000000000" pattern="[A-Fa-f0-9]{32}" required={!this.state.node.isABP} value={this.state.node.appKey || ''} onChange={this.onChange.bind(this, 'appKey')} />
            </div>
          </div>
          <div className={(this.state.activeTab === "advanced-network-settings" ? '' : 'hidden')}>
            <div>
              <p>Please note that changes made below (when updating a node) only have effect after (re)activating the node (either ABP or OTAA).</p>
            </div>
            <div className="form-group">
              <label className="control-label">Receive window</label>
              <div className="radio">
                <label>
                  <input type="radio" name="rxWindow" id="rxWindow1" value="RX1" disabled={this.state.node.useApplicationSettings} checked={this.state.node.rxWindow === "RX1"} onChange={this.onChange.bind(this, 'rxWindow')} /> RX1
                </label>
              </div>
              <div className="radio">
                <label>
                  <input type="radio" name="rxWindow" id="rxWindow2" value="RX2" disabled={this.state.node.useApplicationSettings} checked={this.state.node.rxWindow === "RX2"} onChange={this.onChange.bind(this, 'rxWindow')} /> RX2 (one second after RX1)
                </label>
              </div>
            </div>
            <div className="form-group">
              <label className="control-label">Relax frame-counter</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="relaxFCnt" id="relaxFCnt" disabled={this.state.node.useApplicationSettings} checked={this.state.node.relaxFCnt} onChange={this.onChange.bind(this, 'relaxFCnt')} /> Enable relax frame-counter
                </label>
              </div>
              <p className="help-block">Note that relax frame-counter mode will compromise security as it enables people to perform replay-attacks.</p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rxDelay">Receive window delay</label>
              <input className="form-control" id="rxDelay" type="number" min="0" max="15" disabled={this.state.node.useApplicationSettings} required value={this.state.node.rxDelay || 0} onChange={this.onChange.bind(this, 'rxDelay')} />
              <p className="help-block">The delay in seconds (0-15) between the end of the TX uplink and the opening of the first reception slot (0=1 sec, 1=1 sec, 2=2 sec, 3=3 sec, ... 15=15 sec).</p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rx1DROffset">RX1 data-rate offset</label>
              <input className="form-control" id="rx1DROffset" type="number" disabled={this.state.node.useApplicationSettings} required value={this.state.node.rx1DROffset || 0} onChange={this.onChange.bind(this, 'rx1DROffset')} />
              <p className="help-block">
                Sets the offset between the uplink data rate and the downlink data-rate used to communicate with the end-device on the first reception slot (RX1).
                Please refer to the LoRaWAN specs for the values that are valid in your region.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rx2DR">RX2 data-rate</label>
              <input className="form-control" id="rx2DR" type="number" disabled={this.state.node.useApplicationSettings} required value={this.state.node.rx2DR || 0} onChange={this.onChange.bind(this, 'rx2DR')} />
              <p className="help-block">
                The data-rate to use when RX2 is used as receive window.
                Please refer to the LoRaWAN specs for the values that are valid in your region.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="channelListID">Channel-list</label>
              <select className="form-control" id="channelListID" name="channelListID" disabled={this.state.node.useApplicationSettings} value={this.state.node.channelListID} onChange={this.onChange.bind(this, "channelListID")}>
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
            <div className="form-group">
              <label className="control-label" htmlFor="adrInterval">ADR interval</label>
              <input className="form-control" id="adrInterval" type="number" disabled={this.state.node.useApplicationSettings} required value={this.state.node.adrInterval || 0} onChange={this.onChange.bind(this, 'adrInterval')} />
              <p className="help-block">
                The interval (of frames) after which the network-server will ask the node to change data-rate and / or TX power
                if it can change to a better data-rate or lower TX power. Setting this to 0 will disable ADR. 
              </p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="installationMargin">Installation margin (dB)</label>
              <input className="form-control" id="installationMargin" type="number" disabled={this.state.node.useApplicationSettings} required value={this.state.node.installationMargin || 0} onChange={this.onChange.bind(this, 'installationMargin')} />
              <p className="help-block">
                The installation margin which is taken into account when calculating the ideal data-rate and TX power.
                A higher margin will lower the data-rate, a lower margin will increase the data-rate and possibly packet loss.
                5dB is the default recommended value.
              </p>
            </div>
          </div>
          <hr />
          <button type="submit" className={"btn btn-primary pull-right " + (this.state.disabled ? 'hidden' : '')}>Submit</button>
        </fieldset>
        </form>
      </div>
    );
  }
}

export default NodeForm;
