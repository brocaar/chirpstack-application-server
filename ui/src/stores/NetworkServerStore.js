import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class NetworkServerStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/networkServer.swagger.json", sessionStore.getClientOpts());
  }

  create(networkServer, callbackFunc) {
    this.swagger.then(client => {
      client.apis.NetworkServerService.Create({
        body: {
          networkServer: networkServer,
        },
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.stopLoader();
        this.notifiy("created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.NetworkServerService.Get({
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

  update(networkServer, callbackFunc) {
    this.swagger.then(client => {
      client.apis.NetworkServerService.Update({
        "network_server.id": networkServer.id,
        body: {
          networkServer: networkServer,
        },
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.stopLoader();
        this.notifiy("updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  notifiy(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "network-server has been " + action,
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

  delete(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.NetworkServerService.Delete({
        id: id,
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.stopLoader();
        this.notifiy("deleted");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(organizationID, limit, offset, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.NetworkServerService.List({
        organizationID: organizationID,
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

}

const networkServerStore = new NetworkServerStore();
export default networkServerStore;
window.test = networkServerStore;
