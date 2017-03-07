import { EventEmitter } from "events";
import "whatwg-fetch";
import tokenStore from "./TokenStore";
import { checkStatus, errorHandler } from "./helpers";


class ApplicationStore extends EventEmitter {
  getAll(callbackFunc) {
    fetch("/api/applications?limit=999", {headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if(typeof(responseData.result) === "undefined") {
          callbackFunc([]);
        } else {
          callbackFunc(responseData.result);
        }
      })
      .catch(errorHandler);
  }

  getApplication(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID, {headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createApplication(application, callbackFunc) {
    fetch("/api/applications", {method: "POST", body: JSON.stringify(application), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateApplication(applicationID, application, callbackFunc) {
    fetch("/api/applications/"+applicationID, {method: "PUT", body: JSON.stringify(application), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteApplication(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID, {method: "DELETE", headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const applicationStore = new ApplicationStore();

export default applicationStore;
