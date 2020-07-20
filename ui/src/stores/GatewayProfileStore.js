import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class GatewayProfileStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/gatewayProfile.swagger.json", sessionStore.getClientOpts());
  }

  create(gatewayProfile, callbackFunc) {
    this.swagger.then(client => {
      client.apis.GatewayProfileService.Create({
        body: {
          gatewayProfile: gatewayProfile,
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
      client.apis.GatewayProfileService.Get({
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

  update(gatewayProfile, callbackFunc) {
    this.swagger.then(client => {
      client.apis.GatewayProfileService.Update({
        "gateway_profile.id": gatewayProfile.id,
        body: {
          gatewayProfile: gatewayProfile,
        },
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.notify("updated");
        this.stopLoader();
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  delete(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.GatewayProfileService.Delete({
        id: id,
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.notify("deleted");
        this.stopLoader();
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(networkServerID, limit, offset, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.GatewayProfileService.List({
        networkServerID: networkServerID,
        limit: limit,
        offset: offset,
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
        message: "gateway-profile has been " + action,
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

}

const gatewayProfileStore = new GatewayProfileStore();
export default gatewayProfileStore;
