import { EventEmitter } from "events";
import "whatwg-fetch";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";

class OrganizationStore extends EventEmitter {
  getAll(search, pageSize, offset, callbackFunc) {
    fetch("/api/organizations?limit="+pageSize+"&offset="+offset+"&search="+encodeURIComponent(search), {headers: sessionStore.getHeader()})
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

  getOrganization(organizationID, callbackFunc) {
    fetch("/api/organizations/"+organizationID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createOrganization(organization, callbackFunc) {
    fetch("/api/organizations", {method: "POST", body: JSON.stringify(organization), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateOrganization(organizationID, organization, callbackFunc) {
    fetch("/api/organizations/"+organizationID, {method: "PUT", body: JSON.stringify(organization), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteOrganization(organizationID, callbackFunc) {
    fetch("/api/organizations/"+organizationID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getUsers(organizationID, pageSize, offset, callbackFunc) {
    fetch("/api/organizations/"+organizationID+"/users?limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  addUser(organizationID, user, callbackFunc) {
    fetch("/api/organizations/"+organizationID+"/users", {method: "POST", body: JSON.stringify(user), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
    .catch(errorHandler);
  } 

  getUser(organizationID, userID, callbackFunc) {
    fetch("/api/organizations/"+organizationID+"/users/"+userID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  removeUser(organizationID, userID, callbackFunc) {
    fetch("/api/organizations/"+organizationID+"/users/"+userID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.emit("change");
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateUser(organizationID, userID, user, callbackFunc) {
    fetch("/api/organizations/"+organizationID+"/users/"+userID, {method: "PUT", body: JSON.stringify(user), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        this.emit("change");
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const organizationStore = new OrganizationStore();

export default organizationStore;
