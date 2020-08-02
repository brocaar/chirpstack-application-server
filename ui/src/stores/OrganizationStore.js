import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler, startLoader, stopLoader } from "./helpers";
import dispatcher from "../dispatcher";


class OrganizationStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/organization.swagger.json", sessionStore.getClientOpts());
  }

  create(organization, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.Create({
        body: {
          organization: organization,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.emit("create", organization);
        this.notify("created");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  get(id, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.Get({
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

  update(organization, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.Update({
        "organization.id": organization.id,
        body: {
          organization: organization,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.emit("change", organization);
        this.notify("updated");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  delete(id, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.Delete({
        id: id,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        this.emit("delete", id);
        this.notify("deleted");
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  list(search, limit, offset, callbackFunc) {
    startLoader();
    this.swagger.then((client) => {
      client.apis.OrganizationService.List({
        search: search,
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

  addUser(organizationID, user, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.AddUser({
        "organization_user.organization_id": organizationID,
        body: {
          organizationUser: user,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  getUser(organizationID, userID, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.GetUser({
        organization_id: organizationID,
        user_id: userID,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  deleteUser(organizationID, userID, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.DeleteUser({
        organization_id: organizationID,
        user_id: userID,
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  updateUser(organizationUser, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.UpdateUser({
        "organization_user.organization_id": organizationUser.organizationID,
        "organization_user.user_id": organizationUser.userID,
        body: {
          organizationUser: organizationUser,
        },
      })
      .then(stopLoader)
      .then(checkStatus)
      .then(resp => {
        callbackFunc(resp.obj);
      })
      .catch(errorHandler);
    });
  }

  listUsers(organizationID, limit, offset, callbackFunc) {
    startLoader();
    this.swagger.then(client => {
      client.apis.OrganizationService.ListUsers({
        organization_id: organizationID,
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
        message: "organization has been " + action,
      },
    });
  }

}

const organizationStore = new OrganizationStore();
export default organizationStore;
