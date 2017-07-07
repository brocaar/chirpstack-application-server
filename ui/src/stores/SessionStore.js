import { EventEmitter } from "events";
import { errorHandler, checkStatus } from "./helpers";
import dispatcher from "../dispatcher";

var loginErrorHandler = (error) => {
  error.then((data) => {
    dispatcher.dispatch({
      type: "CREATE_ERROR",
      error: data,
    });
  });
};

class SessionStore extends EventEmitter {
  constructor() {
    super();
    this.user = {};
    this.applications = [];
    this.organizations = [];
    this.settings = {};

    if (this.getToken() !== "") {
      this.fetchProfile(() => {});
    } 
  }

  setToken(token) {
    localStorage.setItem("jwt", token);
  }

  getToken() {
    return localStorage.getItem("jwt");
  }

  setOrganizationID(id) {
    localStorage.setItem("organizationID", id);
  }

  getOrganizationID() {
    return localStorage.getItem("organizationID");
  }

  getHeader() {
    if (this.getToken() !== "") {
      return {
        "Grpc-Metadata-Authorization": this.getToken(),
      }
    } else {
      return {}
    }
  }

  login(login, callbackFunc) {
    fetch("/api/internal/login", {method: "POST", body: JSON.stringify(login)})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.setToken(responseData.jwt);
        this.fetchProfile(callbackFunc);
      })
      .catch(loginErrorHandler);
  }

  fetchProfile(callbackFunc) {
    fetch("/api/internal/profile", {headers: this.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.user = responseData.user;

        if (typeof(responseData.applications) !== "undefined") {
          this.applications = responseData.applications;
        } else {
          this.applications = [];
        }

        if (typeof(responseData.organizations) !== "undefined") {
          this.organizations = responseData.organizations;
        } else {
          this.organizations = [];
        }

        if (typeof(responseData.settings) !== "undefined") {
          this.settings = responseData.settings;
        } else {
          this.settings = {};
        }

        this.emit("change");
        callbackFunc();
      })
      .catch(errorHandler);
  }

  logout(callbackFunc) {
    localStorage.setItem("jwt", "");
    localStorage.setItem("organizationID", "");
    this.user = {};
    this.applications = [];
    this.settings = {};
    this.emit("change");
    callbackFunc();
  }

  getUser() {
    return this.user;
  }

  getSetting(key) {
    return this.settings[key];
  }

  isAdmin() {
    return this.user.isAdmin;
  }

  isApplicationAdmin(applicationID) {
    for (let i = 0; i < this.applications.length; i++) {
      if (Number(this.applications[i].applicationID) === Number(applicationID)) {
        return this.applications[i].isAdmin;
      }
    }
    return false;
  }

  isOrganizationAdmin(organizationID) {
    for (let i = 0; i < this.organizations.length; i++) {
      if (Number(this.organizations[i].organizationID) === Number(organizationID)) {
        return this.organizations[i].isAdmin;
      }
    }
    return false;
  }
}

const sessionStore = new SessionStore();
window.sessionStore = sessionStore;

export default sessionStore;
