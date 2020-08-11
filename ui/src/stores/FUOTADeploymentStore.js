import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler, startLoader, stopLoader } from "./helpers";
import dispatcher from "../dispatcher";


class FUOTADeploymentStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/fuotaDeployment.swagger.json", sessionStore.getClientOpts());
  }

  createForDevice(devEUI, fuotaDeployment, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.CreateForDevice({
        dev_eui: devEUI,
        body: {
          fuotaDeployment: fuotaDeployment,
        },
      })
        .then(stopLoader())
        .then(checkStatus)
        .then(resp => {
          this.notify("created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.Get({
        id: id,
      })
        .then(stopLoader())
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  list(filters, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.List(filters)
        .then(stopLoader())
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  listDeploymentDevices(filters, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.ListDeploymentDevices(filters)
        .then(stopLoader())
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getDeploymentDevice(fuotaDeploymentID, devEUI, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.GetDeploymentDevice({
        fuota_deployment_id: fuotaDeploymentID,
        dev_eui: devEUI,
      })
        .then(stopLoader)
        .then(checkStatus)
        .then(resp => {
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
        message: "fuota deployment has been " + action,
      },
    });
  }

  emitReload() {
    this.emit("reload");
  }
}

const fuotaDeploymentStore = new FUOTADeploymentStore();
export default fuotaDeploymentStore;
