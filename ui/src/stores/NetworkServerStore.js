import { EventEmitter } from "events";
import "whatwg-fetch";
import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";


class NetworkServerStore extends EventEmitter {
  getAll(pageSize, offset, callbackFunc) {
    fetch("/api/network-servers?limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  getAllForOrganizationID(organizationID, pageSize, offset, callbackFunc) {
    fetch("/api/network-servers?limit="+pageSize+"&offset="+offset+"&organizationID="+organizationID, {headers: sessionStore.getHeader()})
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

  getNetworkServer(networkServerID, callbackFunc) {
    fetch("/api/network-servers/"+networkServerID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createNetworkServer(networkServer, callbackFunc) {
    fetch("/api/network-servers", {method: "POST", body: JSON.stringify(networkServer), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateNetworkServer(networkServerID, networkServer, callbackFunc) {
    fetch("/api/network-servers/"+networkServerID, {method: "PUT", body: JSON.stringify(networkServer), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteNetworkServer(networkServerID, callbackFunc) {
    fetch("/api/network-servers/"+networkServerID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const networkServerStore = new NetworkServerStore();

export default networkServerStore;