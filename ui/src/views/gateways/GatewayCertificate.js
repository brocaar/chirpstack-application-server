import React, { Component } from "react";

import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";

import GatewayStore from "../../stores/GatewayStore";


class GatewayCertificate extends Component {
  constructor() {
    super();

    this.state = {
      tlsCert: null,
      tlsKey: null,
      buttonDisabled: false,
    };
  }

  requestCertificate = () => {
    this.setState({
      buttonDisabled: true,
    });

    GatewayStore.generateClientCertificate(this.props.match.params.gatewayID, (resp => {
      this.setState({
        tlsKey: resp.tlsKey,
        tlsCert: resp.tlsCert,
      });
    }));
  }

  render() {
    return(
      <Card>
        <CardContent>
          <Typography gutterBottom>
            When required by the network, the gateway needs a client certificate in order to connect to the network.
            This certificate must be configured on the gateway. After generating the certificate, the certificate
            can only be retrieved once.
          </Typography>
          {this.state.tlsCert == null && <Button onClick={this.requestCertificate} disabled={this.state.buttonDisabled}>Generate certificate</Button>}
          {this.state.tlsCert != null && <form>
            <TextField
              id="tlsCert"
              label="TLS certificate"
              margin="normal"
              value={this.state.tlsCert}
              rows={10}
              multiline
              fullWidth
              helperText="Store this as a text-file on your gateway, e.g. named 'cert.pem'"
            />
            <TextField
              id="tlsKey"
              label="TLS key"
              margin="normal"
              value={this.state.tlsKey}
              rows={10}
              multiline
              fullWidth
              helperText="Store this as a text-file on your gateway, e.g. named 'key.pem'"
            />
          </form>}
        </CardContent>
      </Card>
    );
  }
}

export default GatewayCertificate;
