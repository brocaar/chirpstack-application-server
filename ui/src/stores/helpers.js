import dispatcher from "../dispatcher";
import history from '../history';

export function checkStatus(response) {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    throw response.json();
  }
};

export function errorHandler(error) {
  stopLoader();
  if(error.response === undefined) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "error",
        message: error.message,
      },
    });
  } else {
    if (error.response.obj.code === 16) {
      history.push("/login");
    } else {
      dispatcher.dispatch({
        type: "CREATE_NOTIFICATION",
        notification: {
          type: "error",
          message: error.response.obj.error + " (code: " + error.response.obj.code + ")",
        },
      });
    }
  }
};

export function errorHandlerLogin(error) {
  stopLoader();
  if(error.response === undefined) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "error",
        message: error.message,
      },
    });
  } else {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "error",
        message: error.response.obj.error + " (code: " + error.response.obj.code + ")",
      },
    });
  }
};

export function errorHandlerIgnoreNotFound(error) {
  stopLoader();
  if (error.response === undefined) {
    dispatcher.dispatch({
      type: "CREATE_NOTIFICATION",
      notification: {
        type: "error",
        message: error.message,
      },
    });
  } else {
    if (error.response.obj.code === 16 && history.location.pathname !== "/login") {
      history.push("/login");
    } else if (error.response.obj.code !== 5) {
      dispatcher.dispatch({
        type: "CREATE_NOTIFICATION",
        notification: {
          type: "error",
          message: error.response.obj.error + " (code: " + error.response.obj.code + ")",
        },
      });
    }
  }
};

export function errorHandlerIgnoreNotFoundWithCallback(callbackFunc) {
  return function(error) {
    stopLoader()
    if (error.response.obj.code === 5) {
      callbackFunc(null);
    } else {
      errorHandlerIgnoreNotFound(error);
    }
  }
}

export function startLoader() {
  dispatcher.dispatch({
    type: "START_LOADER",
  });
}

export function stopLoader(response) {
  dispatcher.dispatch({
    type: "STOP_LOADER",
  });
  return response;
}
