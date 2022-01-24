import React, { Component } from "react";
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardHeader from "@material-ui/core/CardHeader";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";

import moment from "moment";

import ApplicationStore from "../../../stores/ApplicationStore";


class MQTTCertificate extends Component {
  constructor() {
    super();

    this.state = {
      caCert: null,
      tlsCert: null,
      tlsKey: null,
      buttonDisabled: false,
    };
  }

  requestCertificate = () => {
    this.setState({
      buttonDisabled: true,
    });

    ApplicationStore.generateMQTTIntegrationClientCertificate(this.props.match.params.applicationID, (resp => {
      this.setState({
        caCert: resp.caCert,
        tlsCert: resp.tlsCert,
        tlsKey: resp.tlsKey,
        expiresAt: moment(resp.expiresAt).format("lll"),
      });
    }));
  }

  render() {
    return(
      <Grid container spacing={4}>
        <Grid item xs={12}>
          <Card>
            <CardHeader title="Generate MQTT client certificate" />
            <CardContent>
              <Typography gutterBottom>
                  If required by the network, the MQTT client needs to be configured with a client certificate
                  in order to connect to the MQTT broker to device data. The generated certificate is
                  application specific. After generating the certificate, the certificate
                  can only be retrieved once.
              </Typography>
              {this.state.tlsCert == null && <Button onClick={this.requestCertificate} disabled={this.state.buttonDisabled}>Generate certificate</Button>}
              {this.state.tlsCert != null && <form>
                <TextField
                  id="expiresAt"
                  label="Certificate expires at"
                  margin="normal"
                  value={this.state.expiresAt}
                  helperText="The certificate expires at this date. Make sure to generate and configure a new certificate for your MQTT client before this expiration date."
                  disabled
                  fullWidth
                />
                <TextField
                  id="caCert"
                  label="CA certificate"
                  margin="normal"
                  value={this.state.caCert}
                  rows={10}
                  multiline
                  fullWidth
                  helperText="The CA certificate is to authenticate the certificate of the server."
                />
                <TextField
                  id="tlsCert"
                  label="TLS certificate"
                  margin="normal"
                  value={this.state.tlsCert}
                  rows={10}
                  multiline
                  fullWidth
                />
                <TextField
                  id="tlsKey"
                  label="TLS key"
                  margin="normal"
                  value={this.state.tlsKey}
                  rows={10}
                  multiline
                  fullWidth
                />
              </form>}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    );
  }
}

export default MQTTCertificate;
