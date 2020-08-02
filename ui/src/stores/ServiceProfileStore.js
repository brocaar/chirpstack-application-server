import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler, startLoader, stopLoader} from "./helpers";
import dispatcher from "../dispatcher";


class ServiceProfileStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/serviceProfile.swagger.json", sessionStore.getClientOpts());
  }

  create(serviceProfile, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.ServiceProfileService.Create({
        body: {
          serviceProfile: serviceProfile,
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
      client.apis.ServiceProfileService.Get({
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

  update(serviceProfile, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.ServiceProfileService.Update({
        "service_profile.id": serviceProfile.id,
        body: {
          serviceProfile: serviceProfile,
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
      client.apis.ServiceProfileService.Delete({
        id: id,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.notify("deleted");
        callbackFunc(resp.ojb);
      })
      .catch(errorHandler);
    });
  }

  list(organizationID, limit, offset, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.ServiceProfileService.List({
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

  notify(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "service-profile has been " + action,
      },
    });
  }

}

const serviceProfileStore = new ServiceProfileStore();
export default serviceProfileStore;
