import { EventEmitter } from "events";
import dispatcher from "../dispatcher";

class ErrorStore extends EventEmitter {
  constructor() {
    super();
    this.errors = [];
  }

  getAll() {
    return this.errors;
  }

  createError(error) {
    const id = Date.now();

    this.errors.push({
      id: id,
      error: error,
    });

    this.emit("change");
  }

  deleteError(id) {
    let err = null;

    for(var error of this.errors) {
      if(error.id === id) {
        err = error
      }
    }

    this.errors.splice(this.errors.indexOf(err), 1);
    this.emit("change");
  }

  handleActions(action) {
    switch(action.type) {
      case "CREATE_ERROR": {
        this.createError(action.error);
        break;
      }
      case "DELETE_ERROR": {
        this.deleteError(action.id);
        break;
      }
      default:
        break;
    }
  }
}

const errorStore = new ErrorStore();
dispatcher.register(errorStore.handleActions.bind(errorStore));

export default errorStore;
