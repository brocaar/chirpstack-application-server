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
    this.branding = {};

    this.fetchBranding( () => {} );

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
        "Grpc-Metadata-Authorization": "Bearer " + this.getToken(),
      }
    } else {
      return {}
    }
  }

  getLogo() {
    if (this.branding) {
        return this.branding.logo;
      } else {
        return null;
    }
  }

  getFooter() {
    if (this.branding) {
        return this.branding.footer;
      } else {
        return null;
    }
  }

  getRegistration() {
    if (this.branding) {
      return this.branding.registration;
    } else {
      return null;
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

  fetchBranding(callbackFunc) {
    fetch("/api/internal/branding", {headers: this.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.branding = responseData;
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
