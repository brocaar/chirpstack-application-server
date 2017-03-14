import React, { Component } from 'react';

import ChannelStore from "../stores/ChannelStore";

class ApplicationForm extends Component {
  constructor() {
    super();
    this.state = {
      activeTab: "application",
      application: {},
      channelLists: [],
    };

    ChannelStore.getAllChannelLists((lists) => {
      this.setState({channelLists: lists});
    });

    this.handleSubmit = this.handleSubmit.bind(this);
    this.changeTab = this.changeTab.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      application: nextProps.application,
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

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.application);
  }

  changeTab(e) {
    e.preventDefault();
    this.setState({
      activeTab: e.target.getAttribute('aria-controls'),
    });
  }

  render() {
    return (
      <div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={(this.state.activeTab === "application" ? 'active' : '')}><a onClick={this.changeTab} href="#application" aria-controls="application">Application details</a></li>
          <li role="presentation" className={(this.state.activeTab === "network-settings" ? 'active' : '')}><a onClick={this.changeTab} href="#network-settings" aria-controls="network-settings">Network settings</a></li>
        </ul>
        <hr />
        <form onSubmit={this.handleSubmit}>
          <div className={(this.state.activeTab === "application" ? '' : 'hidden')}>
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
          </div>
          <div className={(this.state.activeTab === "network-settings" ? '' : 'hidden')}>
            <div className="form-group">
              <label className="control-label">Class-C nodes</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="isClassC" id="isClassC" checked={this.state.application.isClassC} onChange={this.onChange.bind(this, 'isClassC')} /> Class-C nodes
                </label>
              </div>
              <p className="help-block">
                When checked, it means that the nodes for this application are operating in Class-C mode (always listening) and that data will be sent directly to the node. <br/>
                In any other case, the data will be sent as soon as a receive window occurs.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label">ABP (activation by personalisation)</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="isABP" id="isABP" checked={this.state.application.isABP} onChange={this.onChange.bind(this, 'isABP')} /> ABP activation
                </label>
              </div>
              <p className="help-block">When checked, it means that the nodes for this application will be manually activated and that over-the-air activation (OTAA) will be disabled.</p>
            </div>
            <div className="form-group">
              <label className="control-label">Receive window</label>
              <div className="radio">
                <label>
                  <input type="radio" name="rxWindow" id="rxWindow1" value="RX1" checked={this.state.application.rxWindow === "RX1"} onChange={this.onChange.bind(this, 'rxWindow')} /> RX1
                </label>
              </div>
              <div className="radio">
                <label>
                  <input type="radio" name="rxWindow" id="rxWindow2" value="RX2" checked={this.state.application.rxWindow === "RX2"} onChange={this.onChange.bind(this, 'rxWindow')} /> RX2 (one second after RX1)
                </label>
              </div>
            </div>
            <div className="form-group">
              <label className="control-label">Relax frame-counter</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="relaxFCnt" id="relaxFCnt" checked={this.state.application.relaxFCnt} onChange={this.onChange.bind(this, 'relaxFCnt')} /> Enable relax frame-counter
                </label>
              </div>
              <p className="help-block">Note that relax frame-counter mode will compromise security as it enables people to perform replay-attacks.</p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rxDelay">Receive window delay</label>
              <input className="form-control" id="rxDelay" type="number" min="0" max="15" required value={this.state.application.rxDelay || 0} onChange={this.onChange.bind(this, 'rxDelay')} />
              <p className="help-block">The delay in seconds (0-15) between the end of the TX uplink and the opening of the first reception slot (0=1 sec, 1=1 sec, 2=2 sec, 3=3 sec, ... 15=15 sec).</p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rx1DROffset">RX1 data-rate offset</label>
              <input className="form-control" id="rx1DROffset" type="number" required value={this.state.application.rx1DROffset || 0} onChange={this.onChange.bind(this, 'rx1DROffset')} />
              <p className="help-block">
                Sets the offset between the uplink data rate and the downlink data-rate used to communicate with the end-device on the first reception slot (RX1).
                Please refer to the LoRaWAN specs for the values that are valid in your region.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="rx2DR">RX2 data-rate</label>
              <input className="form-control" id="rx2DR" type="number" required value={this.state.application.rx2DR || 0} onChange={this.onChange.bind(this, 'rx2DR')} />
              <p className="help-block">
                The data-rate to use when RX2 is used as receive window.
                Please refer to the LoRaWAN specs for the values that are valid in your region.
              </p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="channelListID">Channel-list</label>
              <select className="form-control" id="channelListID" name="channelListID" value={this.state.application.channelListID} onChange={this.onChange.bind(this, "channelListID")}>
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
              <input className="form-control" id="adrInterval" type="number" required value={this.state.application.adrInterval || 0} onChange={this.onChange.bind(this, 'adrInterval')} />
              <p className="help-block">
                The interval (of frames) after which the network-server will ask the nodes for this application to change data-rate and / or TX power
                if it can change to a better data-rate or lower TX power. Setting this to 0 will disable ADR. 
              </p>
            </div>
            <div className="form-group">
              <label className="control-label" htmlFor="installationMargin">Installation margin (dB)</label>
              <input className="form-control" id="installationMargin" type="number" required value={this.state.application.installationMargin || 0} onChange={this.onChange.bind(this, 'installationMargin')} />
              <p className="help-block">
                The installation margin which is taken into account when calculating the ideal data-rate and TX power.
                A higher margin will lower the data-rate, a lower margin will increase the data-rate and possibly packet loss.
                5dB is the default recommended value.
              </p>
            </div>
          </div>
          <hr />
          <button type="submit" className="btn btn-primary pull-right">Submit</button>
        </form>
      </div>
    );
  }
}

export default ApplicationForm;
