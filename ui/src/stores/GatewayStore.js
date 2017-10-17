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

  getAllForOrganization(organizationID, pageSize, offset, callbackFunc) {
    fetch("/api/gateways?organizationID="+organizationID+"&limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  generateGatewayToken(mac, callbackFunc) {
    fetch("/api/gateways/"+mac+"/token", {method: "POST", headers: sessionStore.getHeader()})
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
        this.emit("change");
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createChannelConfiguration(conf, callbackFunc) {
    fetch("/api/gateways/channelconfigurations", {method: "POST", body: JSON.stringify(conf), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getChannelConfiguration(id, callbackFunc) {
    fetch("/api/gateways/channelconfigurations/"+id, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateChannelConfiguration(id, conf, callbackFunc) {
    fetch("/api/gateways/channelconfigurations/"+id, {method: "PUT", body: JSON.stringify(conf), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteChannelConfiguration(id, callbackFunc) {
    fetch("/api/gateways/channelconfigurations/"+id, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getAllChannelConfigurations(callbackFunc) {
    fetch("/api/gateways/channelconfigurations", {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if (typeof(responseData.result) === "undefined") {
          callbackFunc([]);
        } else {
          callbackFunc(responseData.result);
        }
      })
      .catch(errorHandler);
  }

  createExtraChannel(chan, callbackFunc) {
    fetch("/api/gateways/extrachannels", {method: "POST", body: JSON.stringify(chan), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateExtraChannel(id, chan, callbackFunc) {
    fetch("/api/gateways/extrachannels/"+id, {method: "PUT", body: JSON.stringify(chan), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteExtraChannel(id, callbackFunc) {
    fetch("/api/gateways/extrachannels/"+id, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  getExtraChannelsForChannelConfigurationID(id, callbackFunc) {
    fetch("/api/gateways/channelconfigurations/"+id+"/extrachannels", {headers: sessionStore.getHeader()}) 
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if (typeof(responseData.result) === "undefined") {
          callbackFunc([]);
        } else {
          callbackFunc(responseData.result);
        }
      })
      .catch(errorHandler);
  }

  getLastPing(mac, callbackFunc) {
    fetch("/api/gateways/"+mac+"/pings/last", {headers: sessionStore.getHeader()})
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
