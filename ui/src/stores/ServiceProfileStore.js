import { EventEmitter } from "events";
import "whatwg-fetch";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";


class ServiceProfileStore extends EventEmitter {
  getAllForOrganizationID(organizationID, pageSize, offset, callbackFunc) {
    fetch("/api/service-profiles?organizationID="+organizationID+"&limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  getServiceProfile(serviceProfileID, callbackFunc) {
    fetch("/api/service-profiles/"+serviceProfileID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createServiceProfile(serviceProfile, callbackFunc) {
    fetch("/api/service-profiles", {method: "POST", body: JSON.stringify(serviceProfile), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateServiceProfile(serviceProfileID, serviceProfile, callbackFunc) {
    fetch("/api/service-profiles/"+serviceProfileID, {method: "PUT", body: JSON.stringify(serviceProfile), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteServiceProfile(serviceProfileID, callbackFunc) {
    fetch("/api/service-profiles/"+serviceProfileID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const serviceProfileStore = new ServiceProfileStore();

export default serviceProfileStore;