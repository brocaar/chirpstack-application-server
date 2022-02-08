import { EventEmitter } from "events";
import RobustWebSocket from "robust-websocket";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import { checkStatus, errorHandler, errorHandlerIgnoreNotFoundWithCallback } from "./helpers";
import dispatcher from "../dispatcher";

class DeviceStore extends EventEmitter {
  constructor() {
    super();
    this.wsDataStatus = null;
    this.wsFramesStatus = null;
    this.swagger = new Swagger("/swagger/device.swagger.json", sessionStore.getClientOpts());
  }

  getWSDataStatus() {
    return this.wsDataStatus;
  }

  getWSFramesStatus() {
    return this.wsFramesStatus;
  }

  create(device, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.Create({
        body: {
          device: device,
        },
      })
        .then(checkStatus)
        .then((resp) => {
          this.notify("created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.Get({
        dev_eui: id,
      })
        .then(checkStatus)
        .then((resp) => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  update(device, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.Update({
        "device.dev_eui": device.devEUI,
        body: {
          device: device,
        },
      })
        .then(checkStatus)
        .then((resp) => {
          this.emit("update");
          this.notify("updated");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  delete(id, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.Delete({
        dev_eui: id,
      })
        .then(checkStatus)
        .then((resp) => {
          this.notify("deleted");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  list(filters, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.List(filters)
        .then(checkStatus)
        .then((resp) => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getKeys(devEUI, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.GetKeys({
        dev_eui: devEUI,
      })
        .then(checkStatus)
        .then((resp) => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandlerIgnoreNotFoundWithCallback(callbackFunc));
    });
  }

  createKeys(deviceKeys, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.CreateKeys({
        "device_keys.dev_eui": deviceKeys.devEUI,
        body: {
          deviceKeys: deviceKeys,
        },
      })
        .then(checkStatus)
        .then((resp) => {
          this.notifyKeys("created");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  updateKeys(deviceKeys, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.UpdateKeys({
        "device_keys.dev_eui": deviceKeys.devEUI,
        body: {
          deviceKeys: deviceKeys,
        },
      })
        .then(checkStatus)
        .then((resp) => {
          this.notifyKeys("updated");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getActivation(devEUI, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.GetActivation({
        dev_eui: devEUI,
      })
        .then(checkStatus)
        .then((resp) => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandlerIgnoreNotFoundWithCallback(callbackFunc));
    });
  }

  activate(deviceActivation, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.Activate({
        "device_activation.dev_eui": deviceActivation.devEUI,
        body: {
          deviceActivation: deviceActivation,
        },
      })
        .then(checkStatus)
        .then((resp) => {
          dispatcher.dispatch({
            type: "CREATE_NOTIFICATION",
            notification: {
              type: "success",
              message: "device has been (re)activated",
            },
          });
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getRandomDevAddr(devEUI, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.GetRandomDevAddr({
        dev_eui: devEUI,
      })
        .then(checkStatus)
        .then((resp) => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getStats(devEUI, start, end, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.GetStats({
        dev_eui: devEUI,
        interval: "DAY",
        startTimestamp: start,
        endTimestamp: end,
      })
        .then(checkStatus)
        .then((resp) => {
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  clearDevNonces(id, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.DeviceService.ClearDeviceNonces({
        dev_eui: id,
      })
        .then(checkStatus)
        .then((resp) => {
          this.notifyMsg("device devNonce deleted");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  getDataLogsConnection(devEUI, onData) {
    const loc = window.location;
    const wsURL = (() => {
      if (loc.host === "localhost:3000" || loc.host === "localhost:3001") {
        return `ws://localhost:8080/api/devices/${devEUI}/events`;
      }

      const wsProtocol = loc.protocol === "https:" ? "wss:" : "ws:";
      return `${wsProtocol}//${loc.host}/api/devices/${devEUI}/events`;
    })();

    const conn = new RobustWebSocket(wsURL, ["Bearer", sessionStore.getToken()], {});

    conn.addEventListener("open", () => {
      console.log("connected to", wsURL);
      this.wsDataStatus = "CONNECTED";
      this.emit("ws.status.change");
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
      console.log("closing", wsURL);
      this.wsDataStatus = null;
      this.emit("ws.status.change");
    });

    conn.addEventListener("error", () => {
      console.log("error");
      this.wsDataStatus = "ERROR";
      this.emit("ws.status.change");
    });

    return conn;
  }

  getFrameLogsConnection(devEUI, onData) {
    const loc = window.location;
    const wsURL = (() => {
      if (loc.host === "localhost:3000" || loc.host === "localhost:3001") {
        return `ws://localhost:8080/api/devices/${devEUI}/frames`;
      }

      const wsProtocol = loc.protocol === "https:" ? "wss:" : "ws:";
      return `${wsProtocol}//${loc.host}/api/devices/${devEUI}/frames`;
    })();

    const conn = new RobustWebSocket(wsURL, ["Bearer", sessionStore.getToken()], {});

    conn.addEventListener("open", () => {
      console.log("connected to", wsURL);
      this.wsFramesStatus = "CONNECTED";
      this.emit("ws.status.change");
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
      console.log("closing", wsURL);
      this.wsFramesStatus = null;
      this.emit("ws.status.change");
    });

    conn.addEventListener("error", (e) => {
      console.log("error", e);
      this.wsFramesStatus = "ERROR";
      this.emit("ws.status.change");
    });

    return conn;
  }

  notify(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "device has been " + action,
      },
    });
  }

  notifyKeys(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "device-keys have been " + action,
      },
    });
  }

  notifyMsg(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: action,
      },
    });
  }
}

const deviceStore = new DeviceStore();
export default deviceStore;
