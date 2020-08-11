import { EventEmitter } from "events";
import dispatcher from "../dispatcher";


class LoaderStore extends EventEmitter {
  constructor() {
    super();
    this.totalRequest = 0;
  }

  getTotalRequest() {
    return this.totalRequest;
  }

  incrementTotalRequest() {
    this.totalRequest++;
  }

  decrementTotalRequest() {
    this.totalRequest--;
  }

  handleActions(action) {
    switch(action.type) {
      case "START_LOADER": {
        this.incrementTotalRequest();
        this.emit("change");
        break;
      }
      case "STOP_LOADER": {
        this.decrementTotalRequest();
        this.emit("change");
        break;
      }
      default:
        break;
    }
  }

}


const loaderStore = new LoaderStore();
dispatcher.register(loaderStore.handleActions.bind(loaderStore));

export default loaderStore;
