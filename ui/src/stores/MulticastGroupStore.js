import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class MulticastGroupStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/multicastGroup.swagger.json", sessionStore.getClientOpts());
  }

  create(multicastGroup, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.Create({
        body: {
          multicastGroup: multicastGroup,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.notify("created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.Get({
        id: id,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  update(multicastGroup, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.Update({
        "multicast_group.id": multicastGroup.id,
        body: {
          multicastGroup: multicastGroup,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.notify("updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  delete(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.Delete({
        id: id,
      })
      .then(checkStatus)
      .then(resp => {
        this.notify("deleted");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(search, applicationID, serviceProfileID, devEUI, limit, offset, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.List({
        limit: limit,
        offset: offset,
        applicationID: applicationID,
        serviceProfileID: serviceProfileID,
        devEUI: devEUI,
        search: search,
      })
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  addDevice(multicastGroupID, devEUI, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.AddDevice({
        multicast_group_id: multicastGroupID,
        body: {
          devEUI: devEUI,
        },
      })
      .then(checkStatus)
      .then(resp => {
        this.notifyDevice("added to");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  removeDevice(multicastGroupID, devEUI, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.RemoveDevice({
        multicast_group_id: multicastGroupID,
        dev_eui: devEUI,
      })
      .then(checkStatus)
      .then(resp => {
        this.notifyDevice(devEUI + " removed from");
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
        message: "multicast-group has been " + action,
      },
    });
  }

  notifyDevice(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "device has been " + action + " multicast-group",
      },
    });
  }
}


const multicastGroupStore = new MulticastGroupStore();
export default multicastGroupStore;
