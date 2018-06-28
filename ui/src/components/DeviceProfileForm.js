import React, { Component } from 'react';
import { Link, withRouter } from 'react-router-dom';

import Select from "react-select";

import Loaded from "./Loaded.js";
import NetworkServerStore from "../stores/NetworkServerStore";
import SessionStore from "../stores/SessionStore";


class DeviceProfileForm extends Component {
  constructor() {
    super();

    this.state = {
      deviceProfile: {},
      networkServers: [],
      update: false,
      activeTab: "general",
      isAdmin: false,
      loaded: {
        networkServers: false,
      },
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.changeTab = this.changeTab.bind(this);
  }

  componentDidMount() {
    NetworkServerStore.getAllForOrganizationID(this.props.organizationID, 9999, 0, (totalCount, networkServers) => {
      this.setState({
        deviceProfile: this.props.deviceProfile,
        networkServers: networkServers,
        isAdmin: SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.organizationID),
        loaded: {
          networkServers: true,
        },
      });
    });
  }

  componentWillReceiveProps(nextProps) {
    let dp = nextProps.deviceProfile;
    if (dp !== undefined && dp.factoryPresetFreqs !== undefined && dp.factoryPresetFreqs.length > 0) {
      dp.factoryPresetFreqsStr = dp.factoryPresetFreqs.join(', ');
    }

    this.setState({
      deviceProfile: dp,
      update: nextProps.deviceProfile.id !== undefined,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.deviceProfile);
  }

  onChange(field, e) {
    let deviceProfile = this.state.deviceProfile;

    if (field === "factoryPresetFreqsStr") {
      deviceProfile[field] = e.target.value;

      if (e.target.value === "") {
        deviceProfile[field] = [];
      } else {
        let freqsStr = e.target.value.split(",");
        deviceProfile[field] = freqsStr.map((c, i) => parseInt(c, 10));
      }
    } else if (e.target.type === "number") {
      deviceProfile[field] = parseInt(e.target.value, 10);
    } else if (e.target.type === "checkbox") {
      deviceProfile[field] = e.target.checked;
    } else {
      deviceProfile[field] = e.target.value;
    }

    this.setState({
      deviceProfile: deviceProfile,
    });
  }

  onSelectChange(field, val) {
    let deviceProfile = this.state.deviceProfile;
    deviceProfile[field] = val.value;

    this.setState({
      deviceProfile: deviceProfile,
    });
  }

  changeTab(e) {
    e.preventDefault();
    this.setState({
      activeTab: e.target.getAttribute("aria-controls"),
    });
  }

  render() {
    const networkServerOptions = this.state.networkServers.map((networkServer, i) => {
      return {
        value: networkServer.id,
        label: networkServer.name,
      };
    });

    const macVersionOptions = [
      {value: "1.0.0", label: "1.0.0"},
      {value: "1.0.1", label: "1.0.1"},
      {value: "1.0.2", label: "1.0.2"},
      {value: "1.1.0", label: "1.1.0"},
    ];

    const regParamsOptions = [
      {value: "A", label: "A"},
      {value: "B", label: "B"},
    ];

    const pingSlotPeriodOptions = [
      {value: 32 * 1, label: "every second"},
      {value: 32 * 2, label: "every 2 seconds"},
      {value: 32 * 4, label: "every 4 seconds"},
      {value: 32 * 8, label: "every 8 seconds"},
      {value: 32 * 16, label: "every 16 seconds"},
      {value: 32 * 32, label: "every 32 seconds"},
      {value: 32 * 64, label: "every 64 seconds"},
      {value: 32 * 128, label: "every 128 seconds"},
    ];

    return(
      <Loaded loaded={this.state.loaded}>
        <div>
          <ul className="nav nav-tabs">
            <li role="presentation" className={(this.state.activeTab === "general" ? "active" : "")}><a onClick={this.changeTab} href="#general" aria-controls="general">General</a></li>
            <li role="presentation" className={(this.state.activeTab === "join" ? "active" : "")}><a onClick={this.changeTab} href="#join" aria-controls="join">Join (OTAA / ABP)</a></li>
            <li role="presentation" className={(this.state.activeTab === "classB" ? "active" : "")}><a onClick={this.changeTab} href="#classB" aria-controls="classB">Class-B</a></li>
            <li role="presentation" className={(this.state.activeTab === "classC" ? "active" : "")}><a onClick={this.changeTab} href="#classC" aria-controls="classC">Class-C</a></li>
          </ul>
          <hr />
          <form onSubmit={this.handleSubmit}>
            <div className={"alert alert-warning " + (this.state.networkServers.length > 0 ? 'hidden' : '')}>
              No network-servers are associated with this organization, a <Link to={`/organizations/${this.props.organizationID}/service-profiles`}>service-profile</Link> needs to be created first for this organization.
            </div>
            <fieldset disabled={!this.state.isAdmin}>
              <div className={(this.state.activeTab === "general" ? "" : "hidden")}>
                <div className="form-group">
                  <label className="control-label" htmlFor="name">Device-profile name</label>
                  <input className="form-control" id="name" type="text" placeholder="e.g. my device-profile" required value={this.state.deviceProfile.name || ''} onChange={this.onChange.bind(this, 'name')} />
                  <p className="help-block">
                    A memorable name for the device-profile.
                  </p>
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="networkServerID">Network-server</label>
                  <Select
                    name="networkServerID"
                    options={networkServerOptions}
                    value={this.state.deviceProfile.networkServerID}
                    onChange={this.onSelectChange.bind(this, 'networkServerID')}
                    disabled={this.state.update}
                  />
                  <p className="help-block">
                    The network-server on which this device-profile will be provisioned. After creating the device-profile, this value can't be changed.
                  </p>
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="macVersion">LoRaWAN MAC version</label>
                  <Select 
                    name="macVersion"
                    options={macVersionOptions}
                    value={this.state.deviceProfile.macVersion}
                    onChange={this.onSelectChange.bind(this, 'macVersion')}
                  />
                  <p className="help-block">
                    Version of the LoRaWAN supported by the End-Device.
                  </p>
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="macVersion">LoRaWAN Regional Parameters revision</label>
                  <Select 
                    name="regParamsRevision"
                    options={regParamsOptions}
                    value={this.state.deviceProfile.regParamsRevision}
                    onChange={this.onSelectChange.bind(this, 'regParamsRevision')}
                  />
                  <p className="help-block">
                    Revision of the Regional Parameters document supported by the End-Device.
                  </p>
                </div>
                <div className="form-group">
                  <label className="control-label" htmlFor="maxEIRP">Max EIRP</label>
                  <input className="form-control" name="maxEIRP" id="maxEIRP" type="number" value={this.state.deviceProfile.maxEIRP || 0} onChange={this.onChange.bind(this, 'maxEIRP')} />
                  <p className="help-block">
                    Maximum EIRP supported by the End-Device.
                  </p>
                </div>
              </div>
              <div className={(this.state.activeTab === "join" ? "" : "hidden")}>
                <div className="form-group">
                  <label className="control-label" htmlFor="supportsJoin">Supports join (OTAA)</label>
                  <div className="checkbox">
                    <label>
                      <input type="checkbox" name="supportsJoin" id="supportsJoin" checked={!!this.state.deviceProfile.supportsJoin} onChange={this.onChange.bind(this, 'supportsJoin')} /> Supports join
                    </label>
                  </div>
                  <p className="help-block">
                    End-Device supports Join (OTAA) or not (ABP).
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsJoin === true ? "hidden" : "")}>
                  <label className="control-label" htmlFor="rxDelay1">RX1 delay</label>
                  <input className="form-control" name="rxDelay1" id="rxDelay1" type="number" value={this.state.deviceProfile.rxDelay1 || 0} onChange={this.onChange.bind(this, 'rxDelay1')} />
                  <p className="help-block">
                    Class A RX1 delay (mandatory for ABP).
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsJoin === true ? "hidden" : "")}>
                  <label className="control-label" htmlFor="rxDROffset1">RX1 data-rate offset</label>
                  <input className="form-control" name="rxDROffset1" id="rxDROffset1" type="number" value={this.state.deviceProfile.rxDROffset1 || 0} onChange={this.onChange.bind(this, 'rxDROffset1')} />
                  <p className="help-block">
                    RX1 data rate offset (mandatory for ABP).
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsJoin === true ? "hidden" : "")}>
                  <label className="control-label" htmlFor="rxDataRate2">RX2 data-rate</label>
                  <input className="form-control" name="rxDataRate2" id="rxDataRate2" type="number" value={this.state.deviceProfile.rxDataRate2 || 0} onChange={this.onChange.bind(this, 'rxDataRate2')} />
                  <p className="help-block">
                    RX2 data rate (mandatory for ABP).
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsJoin === true ? "hidden" : "")}>
                  <label className="control-label" htmlFor="rxFreq2">RX2 channel frequency</label>
                  <input className="form-control" name="rxFreq2" id="rxFreq2" type="number" value={this.state.deviceProfile.rxFreq2 || 0} onChange={this.onChange.bind(this, 'rxFreq2')} />
                  <p className="help-block">
                    RX2 channel frequency (mandatory for ABP).
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsJoin === true ? "hidden" : "")}>
                  <label className="control-label" htmlFor="factoryPresetFreqsStr">Factory-present frequencies</label>
                  <input className="form-control" id="factoryPresetFreqsStr" type="text" placeholder="e.g. 868100000, 868300000, 868500000" value={this.state.deviceProfile.factoryPresetFreqsStr || ''} onChange={this.onChange.bind(this, 'factoryPresetFreqsStr')} />
                  <p className="help-block">
                    List of factory-preset frequencies (mandatory for ABP).
                  </p>
                </div>
              </div>
              <div className={(this.state.activeTab === "classB" ? "" : "hidden")}>
                <div className="form-group">
                  <label className="control-label" htmlFor="supportsClassB">Supports Class-B</label>
                  <div className="checkbox">
                    <label>
                      <input type="checkbox" name="supportsClassB" id="supportsClassB" checked={this.state.deviceProfile.supportsClassB} onChange={this.onChange.bind(this, 'supportsClassB')} /> Supports Class-B
                    </label>
                  </div>
                  <p className="help-block">
                    End-Device supports Class B.
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsClassB === true ? "" : "hidden")}>
                  <label className="control-label" htmlFor="classBTimeout">Class-B confirmed downlink timeout</label>
                  <input className="form-control" name="classBTimeout" id="classBTimeout" type="number" value={this.state.deviceProfile.classBTimeout || 0} onChange={this.onChange.bind(this, 'classBTimeout')} />
                  <p className="help-block">
                    Class-B timeout (in seconds) for confirmed downlink transmissions.
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsClassB === true ? "" : "hidden")}>
                  <label className="control-label" htmlFor="pingSlotPeriod">Class-B ping-slot periodicity</label>
                  <Select
                    name="pingSlotPeriod"
                    options={pingSlotPeriodOptions}
                    value={this.state.deviceProfile.pingSlotPeriod}
                    onChange={this.onSelectChange.bind(this, 'pingSlotPeriod')}
                  />
                  <p className="help-block">
                    Class-B ping-slot periodicity.
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsClassB === true ? "" : "hidden")}>
                  <label className="control-label" htmlFor="pingSlotDR">Class-B ping-slot data-rate</label>
                  <input className="form-control" name="pingSlotDR" id="pingSlotDR" type="number" value={this.state.deviceProfile.pingSlotDR || 0} onChange={this.onChange.bind(this, 'pingSlotDR')} />
                  <p className="help-block">
                    Class-B data-rate.
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsClassB === true ? "" : "hidden")}>
                  <label className="control-label" htmlFor="pingSlotFreq">Class-B ping-slot frequency (Hz)</label>
                  <input className="form-control" name="pingSlotFreq" id="pingSlotFreq" type="number" value={this.state.deviceProfile.pingSlotFreq || 0} onChange={this.onChange.bind(this, 'pingSlotFreq')} />
                  <p className="help-block">
                    Class-B frequency (in Hz).
                  </p>
                </div>
              </div>
              <div className={(this.state.activeTab === "classC" ? "" : "hidden")}>
                <div className="form-group">
                  <label className="control-label" htmlFor="supportsClassC">Supports Class-C</label>
                  <div className="checkbox">
                    <label>
                      <input type="checkbox" name="supportsClassC" id="supportsClassC" checked={!!this.state.deviceProfile.supportsClassC} onChange={this.onChange.bind(this, 'supportsClassC')} /> Supports Class-C
                    </label>
                  </div>
                  <p className="help-block">
                    End-Device supports Class C.
                  </p>
                </div>
                <div className={"form-group " + (this.state.deviceProfile.supportsClassC === true ? "" : "hidden")}>
                  <label className="control-label" htmlFor="classCTimeout">Class-C confirmed downlink timeout</label>
                  <input className="form-control" name="classCTimeout" id="classCTimeout" type="number" value={this.state.deviceProfile.classCTimeout || 0} onChange={this.onChange.bind(this, 'classCTimeout')} />
                  <p className="help-block">
                    Class-C timeout (in seconds) for confirmed downlink transmissions.
                  </p>
                </div>
              </div>
            </fieldset>
            <hr />
            <div className={"btn-toolbar pull-right " + (this.state.isAdmin ? "" : "hidden")}>
              <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
              <button type="submit" className="btn btn-primary">Submit</button>
            </div>
          </form>
        </div>
      </Loaded>
    );
  }
}

export default withRouter(DeviceProfileForm);