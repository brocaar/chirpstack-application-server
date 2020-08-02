import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler, startLoader, stopLoader } from "./helpers";
import dispatcher from "../dispatcher";


class DeviceQueueStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/deviceQueue.swagger.json", sessionStore.getClientOpts());
  }

  flush(devEUI, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.DeviceQueueService.Flush({
        dev_eui: devEUI,
      })
        .then(stopLoader)
        .then(checkStatus)
        .then(resp => {
          this.notify("device-queue has been flushed");
          callbackFunc(resp.obj);
        })
        .catch(errorHandler);
    });
  }

  list(devEUI, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.DeviceQueueService.List({
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

  enqueue(item, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.DeviceQueueService.Enqueue({
        "device_queue_item.dev_eui": item.devEUI,
        body: {
          deviceQueueItem: item,
        },
      })
        .then(stopLoader)
        .then(checkStatus)
        .then(resp => {
          this.notify("device-queue item has been created");
          this.emit("enqueue");
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

const deviceQueueStore = new DeviceQueueStore();
export default deviceQueueStore;
