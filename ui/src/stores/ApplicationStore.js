import { EventEmitter } from "events";
import "whatwg-fetch";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";


class ApplicationStore extends EventEmitter {
  getAll(pageSize, offset, callbackFunc) {
    fetch("/api/applications?limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if(typeof(responseData.result) === "undefined") {
          callbackFunc(0, []);
        } else {
          callbackFunc(responseData.totalCount, responseData.result);
        }
      })
      .catch(errorHandler);
  }

  getApplication(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createApplication(application, callbackFunc) {
    fetch("/api/applications", {method: "POST", body: JSON.stringify(application), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateApplication(applicationID, application, callbackFunc) {
    fetch("/api/applications/"+applicationID, {method: "PUT", body: JSON.stringify(application), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteApplication(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getUsers(applicationID, pageSize, offset, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/users?limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if(typeof(responseData.result) === "undefined") {
          callbackFunc(0, []);
        } else {
          callbackFunc(responseData.totalCount, responseData.result);
        }
      })
      .catch(errorHandler);
  }

  addUser(applicationID, user, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/users", {method: "POST", body: JSON.stringify(user), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
    .catch(errorHandler);
  } 

  getUser(applicationID, userID, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/users/"+userID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  removeUser(applicationID, userID, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/users/"+userID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.emit("change");
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateUser(applicationID, userID, user, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/users/"+userID, {method: "PUT", body: JSON.stringify(user), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.emit("change");
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const applicationStore = new ApplicationStore();

export default applicationStore;
