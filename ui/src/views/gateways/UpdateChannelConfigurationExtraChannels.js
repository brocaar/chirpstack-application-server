import React, { Component } from "react";
import Select from 'react-select';

import GatewayStore from "../../stores/GatewayStore";


class ExtraChannel extends Component {
  constructor() {
    super();

    this.onDelete = this.onDelete.bind(this);
  }

  onDelete(e) {
    e.preventDefault();
    if (window.confirm("Are you sure you would like to delete this channel?")) {
      GatewayStore.deleteExtraChannel(this.props.networkServerID, this.props.channel.id, (responseData) => {
        this.props.onChange();
      });
    }
  }

  render() {
    return(
      <tr>
        <td>{this.props.channel.frequency}</td>
        <td>{this.props.channel.modulation}</td>
        <td>{this.props.channel.bandwidth}</td>
        <td>{this.props.channel.spreadFactors.join(", ")}</td>
        <td>{this.props.channel.bitRate > 0 ? this.props.channel.bitRate : ""}</td>
        <td><a className="btn btn-danger btn-xs pull-right" onClick={this.onDelete} href="#delete"><span className="glyphicon glyphicon-remove" aria-hidden="true"></span> remove</a></td>
      </tr>
    );
  }
}


class UpdateChannelConfigurationExtraChannels extends Component {
  constructor() {
    super();

    this.state = {
      extraChannels: [],
    };

    this.onChange = this.onChange.bind(this);
  }

  componentDidMount() {
    this.onChange();
  }

  onChange() {
    GatewayStore.getExtraChannelsForChannelConfigurationID(this.props.match.params.networkServerID, this.props.match.params.channelConfigurationID, (channels) => {
      this.setState({
        extraChannels: channels,
      });
    });
  }

  render() {
    const ExtraChannels = this.state.extraChannels.map((chan, i) => <ExtraChannel key={chan.id} channel={chan} networkServerID={this.props.match.params.networkServerID} onChange={this.onChange} />);

    return(
      <div>
        <div className="panel panel-default">
          <div className="panel-body">
            <table className="table table-hover">
              <thead>
                <tr>
                  <th>Frequency (Hz)</th>
                  <th>Modulation</th>
                  <th>Bandwidth (kHz)</th>
                  <th>Spread-factors (LoRa)</th>
                  <th>Bit rate (FSK)</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {ExtraChannels}
              </tbody>
            </table>
          </div>
        </div>
        <div className="panel panel-default">
          <div className="panel-heading">
            Add extra channel
          </div>
          <div className="panel-body">
            <ExtraChannelForm onSubmit={this.onChange} channelConfigurationID={this.props.match.params.channelConfigurationID} networkServerID={this.props.match.params.networkServerID} />
          </div>
        </div>
      </div>
    );
  }
}

class ExtraChannelForm extends Component {
  constructor() {
    super();

    this.state = {
      channel: {},
    };

    this.onModulationChange = this.onModulationChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  onChange(field, e) {
    let channel = this.state.channel;

    if (field === "spreadFactorsStr") {
      let spreadFactorsStr = e.target.value.split(",");
      channel[field] = e.target.value;
      channel["spreadFactors"] = spreadFactorsStr.map((sf, i) => parseInt(sf, 10));
    } else {
      if (e.target.type === "number") {
        channel[field] = parseInt(e.target.value, 10);  
      } else {
        channel[field] = e.target.value;  
      }
    }

    this.setState({
      channel: channel,
    });
  }

  onModulationChange(val) {
    let channel = this.state.channel;
    channel.modulation = val.value;
    this.setState({
      channel: channel,
    });
  }

  handleSubmit(e) {
    e.preventDefault();

    let channel = this.state.channel;
    channel.channelConfigurationID = this.props.channelConfigurationID;
    channel.networkServerID = this.props.networkServerID;

    GatewayStore.createExtraChannel(channel, (responseData) => {
      this.setState({
        channel: {},
      });

      this.props.onSubmit();
    }); 
  }

  render() {
    const modulations = [
      {value: "LORA", label: "LoRa"},
      {value: "FSK", label: "FSK"},
    ];

    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="modulation">Modulation</label>
          <Select
            name="modulation"
            options={modulations}
            value={this.state.channel.modulation}
            onChange={this.onModulationChange}
            clearable={false}
          />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="frequency">Frequency (Hz)</label>
          <input className="form-control" name="frequency" id="frequency" type="number" placeholder="e.g. 867100000" required value={this.state.channel.frequency || ''} onChange={this.onChange.bind(this, 'frequency')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="bandWidth">Bandwidth (kHz)</label>
          <input className="form-control" name="bandWidth" id="bandWidth" type="number" placeholder="e.g. 125" required value={this.state.channel.bandWidth || ''} onChange={this.onChange.bind(this, 'bandWidth')} />
        </div>
        <div className={"form-group " + (this.state.channel.modulation === "LORA" ? "" : "hidden")}>
          <label className="control-label" htmlFor="spreadFactors">Spread-factors</label>
          <input className="form-control" name="spreadFactors" id="spreadFactorsStr" type="text" placeholder="e.g. 7, 8, 9, 10, 11, 12" required={this.state.channel.modulation === "LORA"} value={this.state.channel.spreadFactorsStr || ''} onChange={this.onChange.bind(this, 'spreadFactorsStr')} />
        </div>
        <div className={"form-group " + (this.state.channel.modulation === "FSK" ? "" : "hidden")}>
          <label className="control-label" htmlFor="bitRate">Bit rate</label>
          <input className="form-control" name="bitRate" id="bitRate" type="number" placeholder="e.g. 50000" required={this.state.channel.modulation === "FSK"} value={this.state.channel.bitRate || ''} onChange={this.onChange.bind(this, 'bitRate')} />
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

export default UpdateChannelConfigurationExtraChannels;
