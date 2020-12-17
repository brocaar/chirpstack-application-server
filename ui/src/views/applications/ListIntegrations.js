import React, { Component } from "react";

import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";

import ApplicationStore from "../../stores/ApplicationStore";
import theme from "../../theme";

import GCPPubSubCard from "./integrations/GCPPubSub";
import HTTPCard from "./integrations/HTTP";
import AzureServiceBusCard from "./integrations/AzureServiceBusCard";
import AWSSNSCard from "./integrations/AWSSNSCard";
import InfluxDBCard from "./integrations/InfluxDBCard";
import ThingsboardCard from "./integrations/ThingsboardCard";
import LoRaCloudCard from "./integrations/LoRaCloudCard";
import MyDevicesCard from "./integrations/MyDevicesCard";
import PilotThingsCard from "./integrations/PilotThingsCard";
import MQTTCard from "./integrations/MQTTCard";


const styles = {
  buttons: {
    textAlign: "right",
  },
  button: {
    marginLeft: 2 * theme.spacing(1),
  },
  icon: {
    marginRight: theme.spacing(1),
  },
};


class ListIntegrations extends Component {
  constructor() {
    super();

    this.state = {
      configured: [],
      available: [],
    };
  }

  componentDidMount() {
    ApplicationStore.on("integration.delete", () => {
      this.loadIntegrations(this.props.match.params.organizationID, this.props.match.params.applicationID);
    });

    this.loadIntegrations(this.props.match.params.organizationID, this.props.match.params.applicationID);
  }

  componentWillUnmount() {
    ApplicationStore.removeAllListeners("integration.delete");
  }

  componentDidUpdate(prevProps) {
    if (this.props === prevProps) {
      return;
    }

    this.loadIntegrations(this.props.match.params.organizationID, this.props.match.params.applicationID);
  }

  loadIntegrations = (organizationID, applicationID) => {
    ApplicationStore.listIntegrations(applicationID, (resp) => {
      let configured = [];
      let available = [];
      const includes = (integrations, kind) => {
        for (let x of integrations) {
          if (x.kind === kind) {
            return true;
          } 
        }

        return false;
      };

      // AWS
      if(includes(resp.result, "AWS_SNS")) {
        configured.push(<AWSSNSCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<AWSSNSCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // Azure
      if(includes(resp.result, "AZURE_SERVICE_BUS")) {
        configured.push(<AzureServiceBusCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<AzureServiceBusCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // GCP
      if(includes(resp.result, "GCP_PUBSUB")) {
        configured.push(<GCPPubSubCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<GCPPubSubCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // HTTP
      if(includes(resp.result, "HTTP")) {
        configured.push(<HTTPCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<HTTPCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // InfluxDB
      if(includes(resp.result, "INFLUXDB")) {
        configured.push(<InfluxDBCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<InfluxDBCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // MyDevices
      if(includes(resp.result, "MYDEVICES")) {
        configured.push(<MyDevicesCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<MyDevicesCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // Global MQTT
      if(includes(resp.result, "MQTT_GLOBAL")) {
        configured.push(<MQTTCard organizationID={organizationID} applicationID={applicationID} />);
      }

      // Pilot Things
      if (includes(resp.result, "PILOT_THINGS")) {
        configured.push(<PilotThingsCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<PilotThingsCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // Semtech LoRa Cloud
      if(includes(resp.result, "LORACLOUD")) {
        configured.push(<LoRaCloudCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<LoRaCloudCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      // Thingsboard
      if(includes(resp.result, "THINGSBOARD")) {
        configured.push(<ThingsboardCard organizationID={organizationID} applicationID={applicationID} />);
      } else {
        available.push(<ThingsboardCard organizationID={organizationID} applicationID={applicationID} add />);
      }

      this.setState({
        configured: configured,
        available: available,
      });
    });
  } 

  render() {
    let configured = this.state.configured.map((row, i) => <Grid key={`configured-${i}`} item xs={6} md={4} xl={3}>{row}</Grid>);
    let available = this.state.available.map((row, i) => <Grid key={`available-${i}`} item xs={6} md={4} xl={3}>{row}</Grid>);

    return(
      <Grid container spacing={4}>
        {configured}
        {available}
      </Grid>
    );
  }
}

export default withStyles(styles)(ListIntegrations);
