import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class ServiceProfileStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/serviceProfile.swagger.json", sessionStore.getClientOpts());
  }

  create(serviceProfile, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ServiceProfileService.Create({
        body: {
          serviceProfile: serviceProfile,
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
      client.apis.ServiceProfileService.Get({
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

  update(serviceProfile, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ServiceProfileService.Update({
        "service_profile.id": serviceProfile.id,
        body: {
          serviceProfile: serviceProfile,
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
      client.apis.ServiceProfileService.Delete({
        id: id,
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.notify("deleted");
        this.stopLoader();
        callbackFunc(resp.ojb);
      })
      .catch(errorHandler);
    });
  }

  list(organizationID, limit, offset, callbackFunc) {
    this.swagger.then(client => {
      client.apis.ServiceProfileService.List({
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

  notify(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "service-profile has been " + action,
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

const serviceProfileStore = new ServiceProfileStore();
export default serviceProfileStore;
