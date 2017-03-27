import { EventEmitter } from "events";

import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";

class GatewayStore extends EventEmitter {
  getAll(pageSize, offset, callbackFunc) {
    fetch("/api/gateways?limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  getGatewayStats(mac, interval, start, end, callbackFunc) {
    fetch("/api/gateways/"+mac+"/stats?interval="+interval+"&startTimestamp="+start+"&endTimestamp="+end, {headers: sessionStore.getHeader()})
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

  getGateway(mac, callbackFunc) {
    fetch("/api/gateways/"+mac, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createGateway(gateway, callbackFunc) {
    fetch("/api/gateways", {method: "POST", body: JSON.stringify(gateway), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteGateway(mac, callbackFunc) {
    fetch("/api/gateways/"+mac, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateGateway(mac, gateway, callbackFunc) {
    fetch("/api/gateways/"+mac, {method: "PUT", body: JSON.stringify(gateway), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const gatewayStore = new GatewayStore();

export default gatewayStore;
