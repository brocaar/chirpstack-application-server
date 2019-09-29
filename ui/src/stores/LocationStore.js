import { EventEmitter } from "events";
import "whatwg-fetch";


class LocationStore extends EventEmitter {
  getLocation(callbackFunc) {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
        callbackFunc(position);
      });
    }
  }
}

const locationStore = new LocationStore();

export default locationStore;
