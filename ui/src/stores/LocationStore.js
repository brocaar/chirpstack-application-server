import { EventEmitter } from "events";
import "whatwg-fetch";
import { checkStatus, errorHandler } from "./helpers";

class LocationStore extends EventEmitter {
  getLocation(callbackFunc) {
    fetch("https://freegeoip.net/json/")
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if(typeof(responseData.latitude) === "undefined") {
          callbackFunc(0, 0);
        } else {
          callbackFunc({ coords: { latitude: responseData.latitude, longitude: responseData.longitude } });
        }
      })
      .catch(errorHandler);
  }

}

const locationStore = new LocationStore();

export default locationStore;
