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
      client.apis.MulticastGroupService.Get({
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

  update(multicastGroup, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.Update({
        "multicast_group.id": multicastGroup.id,
        body: {
          multicastGroup: multicastGroup,
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
      client.apis.MulticastGroupService.Delete({
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

  list(search, organizationID, serviceProfileID, devEUI, limit, offset, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.List({
        limit: limit,
        offset: offset,
        organizationID: organizationID,
        serviceProfileID: serviceProfileID,
        devEUI: devEUI,
        search: search,
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

  addDevice(multicastGroupID, devEUI, callbackFunc) {
    this.swagger.then(client => {
      client.apis.MulticastGroupService.AddDevice({
        multicast_group_id: multicastGroupID,
        body: {
          devEUI: devEUI,
        },
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.notifyDevice("added to");
        this.stopLoader();
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
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.notifyDevice("removed from");
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


const multicastGroupStore = new MulticastGroupStore();
export default multicastGroupStore;
