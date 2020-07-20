import React, { Component } from "react";
import { withRouter } from "react-router-dom";
import { withStyles } from '@material-ui/core/styles';
import blue from "@material-ui/core/colors/blue";
import theme from "../theme";

import LinearProgress from '@material-ui/core/LinearProgress';

import LoaderStore from "../stores/LoaderStore";

const styles = {
  root: {
    zIndex: "3000",
    backgroundColor: blue[600],
    top: 0,
    left: 0,
    right: 0,
    position: "fixed",
    display: "block"
  },
  bar: {
    backgroundColor: theme.palette.common.white,
  },
}

const LoaderProgress = withStyles(styles)(LinearProgress);

class Loader extends Component {
  constructor() {
    super();
    this.state = {
      totalRequest: LoaderStore.getTotalRequest(),
    };
  }

  componentDidMount() {
    LoaderStore.on("change", () => {
      this.setState({
        totalRequest: LoaderStore.getTotalRequest(),
      });
    });
  }

  render() {
    if (this.state.totalRequest !== 0) {
      return(<LoaderProgress variant="indeterminate" />);
    }

    return(null);
  }

}
export default withRouter(Loader);
