import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import { checkStatus, errorHandler, startLoader, stopLoader } from "./helpers";
import dispatcher from "../dispatcher";


class InternalStore extends EventEmitter {
  constructor() {
    super();
    this.client = null;
    this.swagger = Swagger("/swagger/internal.swagger.json", sessionStore.getClientOpts());
  }

  listAPIKeys(filters, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.InternalService.ListAPIKeys(filters)
        .then(stopLoader)
        .then(checkStatus)
        .then(resp => {

          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  deleteAPIKey(id, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.InternalService.DeleteAPIKey({
        id: id,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.notify("api key has been deleted");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  createAPIKey(obj, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.InternalService.CreateAPIKey({
        body: {
          apiKey: obj,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.notify("api key has been created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  settings(callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.InternalService.Settings({})
        .then(stopLoader)
        .then(checkStatus)
        .then(resp => {

          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getDevicesSummary(organizationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.InternalService.GetDevicesSummary({
        organizationID: organizationID,
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getGatewaysSummary(organizationID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.InternalService.GetGatewaysSummary({
        organizationID: organizationID,
      })
        .then(checkStatus)
        .then(resp => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  notify(msg) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: msg,
      },
    });
  }

}

const internalStore = new InternalStore();
export default internalStore;
