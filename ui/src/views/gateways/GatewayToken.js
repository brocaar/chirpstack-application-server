import React, { Component } from 'react';

import GatewayStore from "../../stores/GatewayStore";


class GatewayToken extends Component {
  constructor() {
    super();

    this.state = {
      token: "",
    };

    this.generateToken = this.generateToken.bind(this);
  }

  generateToken() {
    GatewayStore.generateGatewayToken(this.props.match.params.mac, (responseData) => {
      this.setState({
        token: responseData.token,
      });
    });
  }

  render() {
    return(
      <div className="panel panel-default">
        <div className="panel-body">
          <p>
            In order to grant <a href="https://docs.loraserver.io/lora-channel-manager/overview/">LoRa Channel Manager</a> access
            to the gateway API provided by <a href="https://docs.loraserver.io/loraserver/">LoRa Server</a>, a token must be generated.
            Note that this token is specific to this gateway. Generating a new token does not invalidate a previous
            generated token.
          </p>
          <form>
            <div className="form-group">
              <label className="control-label" htmlFor="name">Gateway token</label>
              <input className="form-control" disabled={false} id="name" type="text" value={this.state.token} />
            </div>
            <hr />
            <div className="btn-toolbar pull-right">
              <a className="btn btn-primary" onClick={this.generateToken}>Generate token</a>
            </div>
          </form>
        </div>
      </div>
    );
  }
}

export default GatewayToken;
