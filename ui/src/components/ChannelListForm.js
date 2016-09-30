import React, { Component } from 'react';

class Channel extends Component {
  render() {
    return (
      <div>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Channel #{this.props.channel + 4} frequency (in Hz)</label>
          <input className="form-control" id="name" type="number" min="0" required value={this.props.freqency} onChange={this.props.onChange} />
        </div>
      </div>
    );
  }
}

class ChannelListForm extends Component {
  constructor() {
    super();

    this.state = {
      list: {
        channels: [],
      },
    };
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      list: nextProps.list,
    }); 
  }

  onChange(field, e) {
    let list = this.state.list;
    list[field] = e.target.value;
    this.setState({list: list});
  };

  onFrequencyChange(i, e) {
    let list = this.state.list;
    list.channels[i] = parseInt(e.target.value, 10);
    this.setState({list: list});
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.list);
  }

  render() {
    let channels = this.state.list.channels;
    for(var i = 0; 5 - channels.length; i++) {
      channels.push(0);
    }

    const Channels = this.state.list.channels.map((freq, i) => <Channel key={i} channel={i} freqency={freq} onChange={this.onFrequencyChange.bind(this, i)} />); 

    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Channel-list name</label>
          <input className="form-control" id="name" type="text" required value={this.state.list.name || ''} onChange={this.onChange.bind(this, 'name')} />
        </div>
        {Channels}
        <hr />
        <button type="submit" className="btn btn-primary pull-right">Submit</button>
      </form>
    );
  }
}

export default ChannelListForm;
