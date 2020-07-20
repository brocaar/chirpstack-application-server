import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class OrganizationStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/organization.swagger.json", sessionStore.getClientOpts());
  }

  create(organization, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.Create({
        body: {
          organization: organization,
        },
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.emit("create", organization);
        this.notify("created");
        this.stopLoader();
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.Get({
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

  update(organization, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.Update({
        "organization.id": organization.id,
        body: {
          organization: organization,
        },
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.emit("change", organization);
        this.notify("updated");
        this.stopLoader();
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  delete(id, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.Delete({
        id: id,
      })
      .then(this.startLoader())
      .then(checkStatus)
      .then(resp => {
        this.emit("delete", id);
        this.notify("deleted");
        this.stopLoader();
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(search, limit, offset, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.OrganizationService.List({
        search: search,
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

  addUser(organizationID, user, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.AddUser({
        "organization_user.organization_id": organizationID,
        body: {
          organizationUser: user,
        },
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

  getUser(organizationID, userID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.GetUser({
        organization_id: organizationID,
        user_id: userID,
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

  deleteUser(organizationID, userID, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.DeleteUser({
        organization_id: organizationID,
        user_id: userID,
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

  updateUser(organizationUser, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.UpdateUser({
        "organization_user.organization_id": organizationUser.organizationID,
        "organization_user.user_id": organizationUser.userID,
        body: {
          organizationUser: organizationUser,
        },
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

  listUsers(organizationID, limit, offset, callbackFunc) {
    this.swagger.then(client => {
      client.apis.OrganizationService.ListUsers({
        organization_id: organizationID,
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
        message: "organization has been " + action,
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

const organizationStore = new OrganizationStore();
export default organizationStore;
