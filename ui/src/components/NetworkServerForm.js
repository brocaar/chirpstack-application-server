import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';


class NetworkServerForm extends Component {
  constructor() {
    super();

    this.state = {
      activeTab: "general",
      networkServer: {},
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.changeTab = this.changeTab.bind(this);
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
    if (e.target.type === "checkbox") {
      networkServer[field] = e.target.checked;
    } else if (e.target.type === "number") {
      networkServer[field] = parseInt(e.target.value, 10);
    } else {
      networkServer[field] = e.target.value;
    }
    this.setState({
      networkServer: networkServer,
    });
  }

  changeTab(e) {
    e.preventDefault();
    this.setState({
      activeTab: e.target.getAttribute("aria-controls"),
    });
  }

  render() {
    return(
      <div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={(this.state.activeTab === "general" ? "active" : "")}><a onClick={this.changeTab} href="#general" aria-controls="general">General</a></li>
          <li role="presentation" className={(this.state.activeTab === "gateway-discovery" ? "active" : "")}><a onClick={this.changeTab} href="#gateway-discovery" aria-controls="gateway-discovery">Gateway discovery</a></li>
          <li role="presentation" className={(this.state.activeTab === "certificates" ? "active" : "")}><a onClick={this.changeTab} href="#certificates" aria-controls="certificates">TLS certificates</a></li>
        </ul>
        <hr />
        <form onSubmit={this.handleSubmit}>
          <div className={(this.state.activeTab === "general" ? "" : "hidden")}>
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
          </div>
          <div className={(this.state.activeTab === "gateway-discovery" ? "" : "hidden")}>
            <div className="form-group">
              <label className="control-label" htmlFor="gatewayDiscoveryEnabled">Enable gateway discovery</label>
              <div className="checkbox">
                <label>
                  <input type="checkbox" name="gatewayDiscoveryEnabled" id="gatewayDiscoveryEnabled" checked={!!this.state.networkServer.gatewayDiscoveryEnabled} onChange={this.onChange.bind(this, 'gatewayDiscoveryEnabled')} /> Enable gateway discovery
                </label>
              </div>
              <p className="help-block">
                Enable the gateway discovery feature for this network-server.
              </p>
            </div>
            <div className={"form-group " + (this.state.networkServer.gatewayDiscoveryEnabled === true ? "" : "hidden")}>
              <label className="control-label" htmlFor="gatewayDiscoveryInterval">Interval (per day)</label>
              <input className="form-control" name="gatewayDiscoveryInterval" id="gatewayDiscoveryInterval" type="number" value={this.state.networkServer.gatewayDiscoveryInterval || 0} onChange={this.onChange.bind(this, 'gatewayDiscoveryInterval')} />
              <p className="help-block">
                The number of gateway discovery 'pings' per day that LoRa App Server will broadcast through each gateway.
              </p>
            </div>
            <div className={"form-group " + (this.state.networkServer.gatewayDiscoveryEnabled === true ? "" : "hidden")}>
              <label className="control-label" htmlFor="gatewayDiscoveryInterval">TX Frequency (Hz)</label>
              <input className="form-control" name="gatewayDiscoveryTXFrequency" id="gatewayDiscoveryTXFrequency" type="number" value={this.state.networkServer.gatewayDiscoveryTXFrequency || 0} onChange={this.onChange.bind(this, 'gatewayDiscoveryTXFrequency')} />
              <p className="help-block">
                The frequency (Hz) used for transmitting the gateway discovery 'pings'.
                Please consult the LoRaWAN Regional Parameters specification for the channels valid for each region.
              </p>
            </div>
            <div className={"form-group " + (this.state.networkServer.gatewayDiscoveryEnabled === true ? "" : "hidden")}>
              <label className="control-label" htmlFor="gatewayDiscoveryDR">TX data-rate</label>
              <input className="form-control" name="gatewayDiscoveryDR" id="gatewayDiscoveryDR" type="number" value={this.state.networkServer.gatewayDiscoveryDR || 0} onChange={this.onChange.bind(this, 'gatewayDiscoveryDR')} />
              <p className="help-block">
                The data-rate used for transmitting the gateway discovery 'pings'.
                Please consult the LoRaWAN Regional Parameters specification for the data-rates valid for each region.
              </p>
            </div>
          </div>
          <div className={(this.state.activeTab === "certificates" ? "" : "hidden")}>
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
          </div>
          <hr />
          <div className="btn-toolbar pull-right">
            <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
            <button type="submit" className="btn btn-primary">Submit</button>
          </div>
        </form>
      </div>
    );
  }
}

export default withRouter(NetworkServerForm);