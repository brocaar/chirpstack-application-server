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

  getAllForOrganization(organizationID, pageSize, offset, callbackFunc) {
    fetch("/api/applications?organizationID="+organizationID+"&limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  createHTTPIntegration(applicationID, integration, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/integrations/http", {method: "POST", body: JSON.stringify(integration), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getHTTPIntegration(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/integrations/http", {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateHTTPIntegration(applicationID, integration, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/integrations/http", {method: "PUT", body: JSON.stringify(integration), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteHTTPIntegration(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/integrations/http", {method: "DELETE", headers: sessionStore.getHeader()}) 
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  listIntegrations(applicationID, callbackFunc) {
    fetch("/api/applications/"+applicationID+"/integrations", {headers: sessionStore.getHeader()}) 
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
