import { EventEmitter } from "events";
import RobustWebSocket from "robust-websocket";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler, errorHandlerIgnoreNotFound, startLoader, stopLoader } from "./helpers";
import dispatcher from "../dispatcher";


class GatewayStore extends EventEmitter {
  constructor() {
    super();
    this.wsStatus = null;
    this.swagger = new Swagger("/swagger/gateway.swagger.json", sessionStore.getClientOpts());
  }

  getWSStatus() {
    return this.wsStatus;
  }

  create(gateway, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayService.Create({
        body: {
          gateway: gateway,
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
      client.apis.GatewayService.Get({
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

  update(gateway, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayService.Update({
        "gateway.id": gateway.id,
        body: {
          gateway: gateway,
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
      client.apis.GatewayService.Delete({
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

  generateClientCertificate(id, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayService.GenerateGatewayClientCertificate({
        gateway_id: id,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(search, organizationID, limit, offset, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayService.List({
        limit: limit,
        offset: offset,
        organizationID: organizationID,
        search: search,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  getStats(gatewayID, start, end, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayService.GetStats({
        gateway_id: gatewayID,
        interval: "DAY",
        startTimestamp: start,
        endTimestamp: end,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  getLastPing(gatewayID, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.GatewayService.GetLastPing({
        gateway_id: gatewayID,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandlerIgnoreNotFound);
    });
  }

  getFrameLogsConnection(gatewayID, onOpen, onClose, onData) {
    const loc = window.location;
    const wsURL = (() => {
      if (loc.host === "localhost:3000" || loc.host === "localhost:3001") {
        return `wss://localhost:8080/api/gateways/${gatewayID}/frames`;
      }

      const wsProtocol = loc.protocol === "https:" ? "wss:" : "ws:";
      return `${wsProtocol}//${loc.host}/api/gateways/${gatewayID}/frames`;
    })();

    const conn = new RobustWebSocket(wsURL, ["Bearer", sessionStore.getToken()], {});

    conn.addEventListener("open", () => {
      console.log('connected to', wsURL);
      this.wsStatus = "CONNECTED";
      this.emit("ws.status.change");
      onOpen();
    });

    conn.addEventListener("message", (e) => {
      const msg = JSON.parse(e.data);
      if (msg.error !== undefined) {
        dispatcher.dispatch({
          type: "CREATE_NOTIFICATION",
          notification: {
            type: "error",
            message: msg.error.message,
          },
        });
      } else if (msg.result !== undefined) {
        onData(msg.result);
      }
    });

    conn.addEventListener("close", () => {
      console.log('closing', wsURL);
      this.wsStatus = null;
      this.emit("ws.status.change");
      onClose();
    });

    conn.addEventListener("error", () => {
      console.log("error");
      this.wsStatus = "ERROR";
      this.emit("ws.status.change");
    });

    return conn;
  }

  notify(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "gateway has been " + action,
      },
    });
  }

}

const gatewayStore = new GatewayStore();
export default gatewayStore;
