import { EventEmitter } from "events";

import Swagger from "swagger-client";

import sessionStore from "./SessionStore";
import {checkStatus, errorHandler } from "./helpers";
import dispatcher from "../dispatcher";


class UserStore extends EventEmitter {
  constructor() {
    super();
    this.swagger = new Swagger("/swagger/user.swagger.json", sessionStore.getClientOpts());
  }

  create(user, password, organizations, callbackFunc) {
    this.swagger.then(client => {
      client.apis.UserService.Create({
        body: {
          organizations: organizations,
          password: password,
          user: user,
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
      client.apis.UserService.Get({
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

  update(user, callbackFunc) {
    this.swagger.then(client => {
      client.apis.UserService.Update({
        "user.id": user.id,
        body: {
          "user": user,
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
      client.apis.UserService.Delete({
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

  updatePassword(id, password, callbackFunc) {
    this.swagger.then(client => {
      client.apis.UserService.UpdatePassword({
        "user_id": id,
        body: {
          password: password,
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

  list(search, limit, offset, callbackFunc) {
    this.swagger.then((client) => {
      client.apis.UserService.List({
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

  notify(action) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "success",
        message: "user has been " + action,
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

const userStore = new UserStore();
export default userStore;
