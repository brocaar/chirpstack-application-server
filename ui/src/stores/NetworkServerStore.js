import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler, startLoader, stopLoader } from "./helpers";
import dispatcher from "../dispatcher";


class NetworkServerStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/networkServer.swagger.json", sessionStore.getClientOpts());
  }

  create(networkServer, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.NetworkServerService.Create({
        body: {
          networkServer: networkServer,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.notifiy("created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    startLoader();
    this.swagger.then((client) => {
      client.apis.NetworkServerService.Get({
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

  update(networkServer, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.NetworkServerService.Update({
        "network_server.id": networkServer.id,
        body: {
          networkServer: networkServer,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
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

  delete(id, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.NetworkServerService.Delete({
        id: id,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.notifiy("deleted");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(organizationID, limit, offset, callbackFunc) {
    startLoader();
    this.swagger.then((client) => {
      client.apis.NetworkServerService.List({
        organizationID: organizationID,
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

}

const networkServerStore = new NetworkServerStore();
export default networkServerStore;
window.test = networkServerStore;
