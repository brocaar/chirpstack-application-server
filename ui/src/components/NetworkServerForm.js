import React, { Component } from 'react';


class NetworkServerForm extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      networkServer: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      networkServer: nextProps.networkServer,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.networkServer);
  }

  onChange(field, e) {
    let networkServer = this.state.networkServer;
    networkServer[field] = e.target.value;
    this.setState({
      networkServer: networkServer,
    });
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Network-server name</label>
          <input className="form-control" id="name" type="text" placeholder="e.g. EU868 network-server" required value={this.state.networkServer.name || ''} onChange={this.onChange.bind(this, 'name')} />
          <p className="help-block">
            A memorable name of the network-server.
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="server">Network-server server</label>
          <input className="form-control" id="server" type="text" placeholder="e.g. localhost:8000" required value={this.state.networkServer.server || ''} onChange={this.onChange.bind(this, 'server')} />
          <p className="help-block">
            The hostname:IP of the network-server.
          </p>
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.context.router.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

export default NetworkServerForm;