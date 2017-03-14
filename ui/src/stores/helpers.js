import dispatcher from "../dispatcher";
import { hashHistory } from "react-router";

export function checkStatus(response) {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    throw response.json();
  }
};

export function errorHandler(error) {
  console.log("error", error);
  error.then((data) => {
    if (data.code === 16) {
      hashHistory.push("/login");
    } else {
      dispatcher.dispatch({
        type: "CREATE_ERROR",
        error: data,
      });
    }
  });
};
