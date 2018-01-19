import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';


class ChannelConfigurationForm extends Component {
  constructor() {
    super();

    this.state = {
      configuration: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentDidMount() {
    let configuration = this.props.configuration;
    if (typeof(configuration.channels) !== "undefined" && typeof(configuration.channelsStr) === "undefined") {
      configuration.channelsStr = configuration.channels.join(", ");
    }

    this.setState({
      configuration: configuration,
    });
  }

  componentWillReceiveProps(nextProps) {
    let configuration = nextProps.configuration;
    if (typeof(configuration.channels) !== "undefined" && typeof(configuration.channelsStr) === "undefined") {
      configuration.channelsStr = configuration.channels.join(", ");
    }

    this.setState({
      configuration: configuration,
    });
  }

  onChange(field, e) {
    let configuration = this.state.configuration;

    if (field === "channelsStr") {
      let channelsStr = e.target.value.split(",");
      configuration[field] = e.target.value;
      configuration["channels"] = channelsStr.map((c, i) => parseInt(c, 10));
    } else {
      configuration[field] = e.target.value;
    }

    this.setState({
      configuration: configuration,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.configuration);
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Configuration name</label>
          <input className="form-control" id="name" type="text" placeholder="a short name identifying the channel-configuration" required value={this.state.configuration.name || ''} onChange={this.onChange.bind(this, 'name')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Enabled channels</label>
          <input className="form-control" id="channelsStr" type="text" placeholder="0, 1, 2" required pattern="[0-9]+(,[\s]*[0-9]+)*" value={this.state.configuration.channelsStr || ''} onChange={this.onChange.bind(this, 'channelsStr')} />
          <p className="help-block">
            The channels active in this channel-configuration as specified in the LoRaWAN Regional Parameters sepecification.
            Channels part of the CFList option can be configured after creating the channel-configuration.
            Separate channels by comma, e.g. 0, 1, 2.
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

export default withRouter(ChannelConfigurationForm);
