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
        <fieldset>
          <legend>Certificates for LoRa App Server to LoRa Server connection</legend>
          <div className="form-group">
            <label className="control-label" htmlFor="caCert">CA certificate</label>
            <textarea className="form-control" rows="4" id="caCert" value={this.state.networkServer.caCert} onChange={this.onChange.bind(this, 'caCert')} />
            <p className="help-block">
              Paste the content of the CA certificate (PEM) file in the above textbox. Leave blank to disable TLS.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="tlsCert">TLS certificate</label>
            <textarea className="form-control" rows="4" id="tlsCert" value={this.state.networkServer.tlsCert} onChange={this.onChange.bind(this, 'tlsCert')} />
            <p className="help-block">
              Paste the content of the TLS certificate (PEM) file in the above textbox. Leave blank to disable TLS.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="tlsKey">TLS key</label>
            <textarea className="form-control" rows="4" id="tlsKey" value={this.state.networkServer.tlsKey} onChange={this.onChange.bind(this, 'tlsKey')} />
            <p className="help-block">
              Paste the content of the TLS key (PEM) file in the above textbox. Leave blank to disable TLS.
              Note: for security reasons, the TLS key can't be retrieved after being submitted (the field is left blank).
              When re-submitting the form with an empty TLS key field (but populated TLS certificate field), the key won't be overwritten.
            </p>
          </div>
        </fieldset>
        <fieldset>
          <legend>Certificates for LoRa Server to LoRa App Server connection</legend>
          <div className="form-group">
            <label className="control-label" htmlFor="routingProfileCACert">CA certificate</label>
            <textarea className="form-control" rows="4" id="routingProfileCACert" value={this.state.networkServer.routingProfileCACert} onChange={this.onChange.bind(this, 'routingProfileCACert')} />
            <p className="help-block">
              Paste the content of the CA certificate (PEM) file in the above textbox. Leave blank to disable TLS.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="routingProfileTLSCert">TLS certificate</label>
            <textarea className="form-control" rows="4" id="routingProfileTLSCert" value={this.state.networkServer.routingProfileTLSCert} onChange={this.onChange.bind(this, 'routingProfileTLSCert')} />
            <p className="help-block">
              Paste the content of the TLS certificate (PEM) file in the above textbox. Leave blank to disable TLS.
            </p>
          </div>
          <div className="form-group">
            <label className="control-label" htmlFor="routingProfileTLSKey">TLS key</label>
            <textarea className="form-control" rows="4" id="routingProfileTLSKey" value={this.state.networkServer.routingProfileTLSKey} onChange={this.onChange.bind(this, 'routingProfileTLSKey')} />
            <p className="help-block">
              Paste the content of the TLS key (PEM) file in the above textbox. Leave blank to disable TLS.
              Note: for security reasons, the TLS key can't be retrieved after being submitted (the field is left blank).
              When re-submitting the form with an empty TLS key field (but populated TLS certificate field), the key won't be overwritten.
            </p>
          </div>
        </fieldset>
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