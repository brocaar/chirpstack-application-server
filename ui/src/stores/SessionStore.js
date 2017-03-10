import { EventEmitter } from "events";
import { checkStatus } from "./helpers";
import dispatcher from "../dispatcher";

var errorHandler = (error) => {
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
      .catch(errorHandler);
  }

  fetchProfile(callbackFunc) {
    fetch("/api/internal/profile", {headers: this.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.user = responseData.user;
        this.applications = responseData.applications;
        this.emit("change");
        callbackFunc();
      })
      .catch(errorHandler);
  }

  logout(callbackFunc) {
    localStorage.setItem("jwt", "");
    this.user = {};
    this.applications = {};
    this.emit("change");
    callbackFunc();
  }

  getUser() {
    return this.user;
  }
}

const sessionStore = new SessionStore();

export default sessionStore;
