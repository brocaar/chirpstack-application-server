import { EventEmitter } from "events";
import "whatwg-fetch";
import dispatcher from "../dispatcher";
import history from "../history";


function checkStatus(response) {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    throw response.json();
  }
};

function errorHandler(error) {
  error.then((data) => {
    if (data.code === 16) {
      history.push("/login");
    } else {
      dispatcher.dispatch({
        type: "CREATE_ERROR",
        error: data,
      });
    }
  });
};


class LocationStore extends EventEmitter {
  getLocation(callbackFunc) {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
        callbackFunc(position);
      }, (error) => {
        this.getGeoIPLocation(callbackFunc);
      });
    } else {
      this.getGeoIPLocation(callbackFunc);
    }
  }

  getGeoIPLocation(callbackFunc) {
    fetch("https://freegeoip.net/json/")
      .then(checkStatus)
      .then((response) => response.json())
      .then((responseData) => {
        if(typeof(responseData.latitude) === "undefined") {
          callbackFunc({coords: {latitude: 0, longitude: 0}});
        } else {
          callbackFunc({ coords: { latitude: responseData.latitude, longitude: responseData.longitude } });
        }
      })
      .catch(errorHandler); 
  }
}

const locationStore = new LocationStore();

export default locationStore;
