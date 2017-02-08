import { EventEmitter } from "events";
import { hashHistory } from "react-router";
import "whatwg-fetch";
import dispatcher from "../dispatcher";
import tokenStore from "./TokenStore";

var checkStatus = (response) => {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    throw response.json();
  }
};

var errorHandler = (error) => {
  error.then((data) => {
    dispatcher.dispatch({
      type: "CREATE_ERROR",
      error: data,
    });

    if (data.Code === 16) {
      hashHistory.push("/jwt");
    }
  });
};

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

  getApplication(applicationName, callbackFunc) {
    fetch("/api/applications/"+applicationName, {headers: tokenStore.getHeader()})
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

  updateApplication(applicationName, application, callbackFunc) {
    fetch("/api/applications/"+applicationName, {method: "PUT", body: JSON.stringify(application), headers: tokenStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteApplication(applicationName, callbackFunc) {
    fetch("/api/applications/"+applicationName, {method: "DELETE", headers: tokenStore.getHeader()})
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
