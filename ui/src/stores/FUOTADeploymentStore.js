import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class FUOTADeploymentStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/fuotaDeployment.swagger.json", sessionStore.getClientOpts());
  }

  createForDevice(devEUI, fuotaDeployment, callbackFunc) {
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.CreateForDevice({
        dev_eui: devEUI,
        body: {
          fuotaDeployment: fuotaDeployment,
        },
      })
        .then(this.startLoader())
        .then(checkStatus)
        .then(resp => {
          this.notify("created");
          this.stopLoader();
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.Get({
        id: id,
      })
        .then(this.startLoader())
        .then(checkStatus)
        .then(resp => {
          this.stopLoader();
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  list(filters, callbackFunc) {
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.List(filters)
        .then(this.startLoader())
        .then(checkStatus)
        .then(resp => {
          this.stopLoader();
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  listDeploymentDevices(filters, callbackFunc) {
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.ListDeploymentDevices(filters)
        .then(this.startLoader())
        .then(checkStatus)
        .then(resp => {
          this.stopLoader();
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getDeploymentDevice(fuotaDeploymentID, devEUI, callbackFunc) {
    this.swagger.then(client => {
      client.apis.FUOTADeploymentService.GetDeploymentDevice({
        fuota_deployment_id: fuotaDeploymentID,
        dev_eui: devEUI,
      })
        .then(this.startLoader())
        .then(checkStatus)
        .then(resp => {
          this.stopLoader();
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

  startLoader() {
    dispatcher.dispatch({
      type: "START_LOADER",
    });
  }

  stopLoader() {
    dispatcher.dispatch({
      type: "STOP_LOADER",
    });
  }

  emitReload() {
    this.emit("reload");
  }
}

const fuotaDeploymentStore = new FUOTADeploymentStore();
export default fuotaDeploymentStore;
