import { EventEmitter } from "events";
import "whatwg-fetch";

import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";


class GatewayProfileStore extends EventEmitter {
  getAllForNetworkServerID(networkServerID, pageSize, offset, callbackFunc) {
    fetch(`/api/gateway-profiles?networkServerID=${networkServerID}&limit=${pageSize}&offset=${offset}`, {headers: sessionStore.getHeader()})
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

  getGatewayProfile(id, callbackFunc) {
    fetch(`/api/gateway-profiles/${id}`, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createGatewayProfile(gatewayProfile, callbackFunc) {
    fetch(`/api/gateway-profiles`, {method: "POST", body: JSON.stringify(gatewayProfile), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateGatewayProfile(gatewayProfileID, gatewayProfile, callbackFunc) {
    fetch(`/api/gateway-profiles/${gatewayProfileID}`, {method: "PUT", body: JSON.stringify(gatewayProfile), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteGatewayProfile(gatewayProfileID, callbackFunc) {
    fetch(`/api/gateway-profiles/${gatewayProfileID}`, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}


const gatewayProfileStore = new GatewayProfileStore();

export default gatewayProfileStore;
