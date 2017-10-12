import { EventEmitter } from "events";
import "whatwg-fetch";

import sessionStore from "./SessionStore";
import { checkStatus, errorHandler } from "./helpers";


class DeviceProfileStore extends EventEmitter {
  getAllForOrganizationID(organizationID, pageSize, offset, callbackFunc) {
    fetch("/api/device-profiles?organizationID="+organizationID+"&limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  getAllForApplicationID(applicationID, pageSize, offset, callbackFunc) {
    fetch("/api/device-profiles?applicationID="+applicationID+"&limit="+pageSize+"&offset="+offset, {headers: sessionStore.getHeader()})
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

  getDeviceProfile(deviceProfileID, callbackFunc) {
    fetch("/api/device-profiles/"+deviceProfileID, {headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  createDeviceProfile(deviceProfile, callbackFunc) {
    fetch("/api/device-profiles", {method: "POST", body: JSON.stringify(deviceProfile), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  updateDeviceProfile(deviceProfileID, deviceProfile, callbackFunc) {
    fetch("/api/device-profiles/"+deviceProfileID, {method: "PUT", body: JSON.stringify(deviceProfile), headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }

  deleteDeviceProfile(deviceProfileID, callbackFunc) {
    fetch("/api/device-profiles/"+deviceProfileID, {method: "DELETE", headers: sessionStore.getHeader()})
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        callbackFunc(responseData);
      })
      .catch(errorHandler);
  }
}

const deviceProfileStore = new DeviceProfileStore();

export default deviceProfileStore;