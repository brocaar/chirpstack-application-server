import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class ApplicationStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/application.swagger.json", sessionStore.getClientOpts());
  }

  create(application, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.Create({
        body: {
          application: application,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.notify("created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.Get({
        id: id,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  update(application, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.Update({
        "application.id": application.id,
        body: {
          application: application,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.notify("updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  delete(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.Delete({
        id: id,
      })
      .then(checkStatus)
      .then(resp => {
        this.notify("deleted");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(search, organizationID, limit, offset, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.List({
        limit: limit,
        offset: offset,
        organizationID: organizationID,
        search: search,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  listIntegrations(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.ListIntegrations({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  createHTTPIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateHTTPIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("http", "created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  getHTTPIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetHTTPIntegration({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  updateHTTPIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateHTTPIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("http", "updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  deleteHTTPIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteHTTPIntegration({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("http", "deleted");
        this.emit("integration.delete");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  createInfluxDBIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateInfluxDBIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("InfluxDB", "created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  getInfluxDBIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetInfluxDBIntegration({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  updateInfluxDBIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateInfluxDBIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("InfluxDB", "updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  deleteInfluxDBIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteInfluxDBIntegration({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("InfluxDB", "deleted");
        this.emit("integration.delete");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  createThingsBoardIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateThingsBoardIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp =>  {
          this.integrationNotification("ThingsBoard.io", "created");
          callbackFunc(resp.obj);
        })
      .catch(errorHandler);
    });
  }

  getThingsBoardIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetThingsBoardIntegration({
        application_id: applicationID, 
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
      .catch(errorHandler);
    });
  }

  updateThingsBoardIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateThingsBoardIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("ThingsBoard.io", "updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  deleteThingsBoardIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteThingsBoardIntegration({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("ThingsBoard.io", "deleted");
        this.emit("integration.delete");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  createMyDevicesIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateMyDevicesIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp =>  {
          this.integrationNotification("myDevices", "created");
          callbackFunc(resp.obj);
        })
      .catch(errorHandler);
    });
  }

  getMyDevicesIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetMyDevicesIntegration({
        application_id: applicationID, 
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
      .catch(errorHandler);
    });
  }

  updateMyDevicesIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateMyDevicesIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("myDevices", "updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  deleteMyDevicesIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteMyDevicesIntegration({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("myDevices", "deleted");
        this.emit("integration.delete");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  createLoRaCloudIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateLoRaCloudIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("LoRa Cloud", "created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getLoRaCloudIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetLoRaCloudIntegration({
        application_id: applicationID, 
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
      .catch(errorHandler);
    });
  }

  updateLoRaCloudIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateLoRaCloudIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("LoRa Cloud", "updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  deleteLoRaCloudIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteLoRaCloudIntegration({
        application_id: applicationID,
      })
      .then(checkStatus)
      .then(resp => {
        this.integrationNotification("LoRa Cloud", "deleted");
        this.emit("integration.delete");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  notify(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "application has been " + action,
      },
    });
  }

  createGCPPubSubIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateGCPPubSubIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("GCP Pub/Sub", "created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getGCPPubSubIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetGCPPubSubIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  updateGCPPubSubIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateGCPPubSubIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("GCP Pub/Sub", "updated");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  deleteGCPPubSubIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteGCPPubSubIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("GCP Pub/Sbu", "deleted");
          this.emit("integration.delete");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  createAWSSNSIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateAWSSNSIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("AWS SNS", "created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getAWSSNSIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetAWSSNSIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  updateAWSSNSIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateAWSSNSIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("AWS SNS", "updated");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  deleteAWSSNSIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteAWSSNSIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("AWS SNS", "deleted");
          this.emit("integration.delete");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  createAzureServiceBusIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreateAzureServiceBusIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("Azure Service-Bus", "created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getAzureServiceBusIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetAzureServiceBusIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  updateAzureServiceBusIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdateAzureServiceBusIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("Azure Service-Bus", "updated");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  deleteAzureServiceBusIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeleteAzureServiceBusIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("Azure Service-Bus", "deleted");
          this.emit("integration.delete");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  createPilotThingsIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.CreatePilotThingsIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("Pilot Things", "created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getPilotThingsIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GetPilotThingsIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  updatePilotThingsIntegration(integration, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.UpdatePilotThingsIntegration({
        "integration.application_id": integration.applicationID,
        body: {
          integration: integration,
        },
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("Pilot Things", "updated");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  deletePilotThingsIntegration(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.DeletePilotThingsIntegration({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          this.integrationNotification("Pilot Things", "deleted");
          this.emit("integration.delete");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  generateMQTTIntegrationClientCertificate(applicationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ApplicationService.GenerateMQTTIntegrationClientCertificate({
        application_id: applicationID,
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  integrationNotification(kind, action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: kind + " integration has been " + action,
      },
    });
  }
}

const applicationStore = new ApplicationStore();
export default applicationStore;
