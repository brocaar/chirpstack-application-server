import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler, startLoader, stopLoader } from "./helpers";
import dispatcher from "../dispatcher";


class GatewayProfileStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/gatewayProfile.swagger.json", sessionStore.getClientOpts());
  }

  create(gatewayProfile, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayProfileService.Create({
        body: {
          gatewayProfile: gatewayProfile,
        },
      })
      .then(stopLoader)
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
      client.apis.GatewayProfileService.Get({
        id: id,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  update(gatewayProfile, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayProfileService.Update({
        "gateway_profile.id": gatewayProfile.id,
        body: {
          gatewayProfile: gatewayProfile,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.notify("updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  delete(id, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayProfileService.Delete({
        id: id,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.notify("deleted");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(networkServerID, limit, offset, callbackFunc) {
    startLoader();
    this.swagger.then((client) => {
      client.apis.GatewayProfileService.List({
        networkServerID: networkServerID,
        limit: limit,
        offset: offset,
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
        message: "gateway-profile has been " + action,
      },
    });
  }

}

const gatewayProfileStore = new GatewayProfileStore();
export default gatewayProfileStore;
