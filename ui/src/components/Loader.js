import React, { Component } from "react";
import { withRouter } from "react-router-dom";
import { withStyles } from '@material-ui/core/styles';
import blue from "@material-ui/core/colors/blue";
import theme from "../theme";

import LinearProgress from '@material-ui/core/LinearProgress';

import LoaderStore from "../stores/LoaderStore";

import moment from "moment";

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
      display: "hide",
      lastLoading: 0,
      status: "stop"
    };
  }

  componentDidMount() {
    LoaderStore.on("change", () => {
      this.setState({
        totalRequest: LoaderStore.getTotalRequest(),
      });
    });
    this.timer = setInterval(this.progress, 10);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  componentDidUpdate(prevProps) {
    if (this.props === prevProps) {
      return;
    }
    this.progress();
  }

  progress = () => {
    const showThreshold = 2;
    if (moment().unix() - this.state.lastLoading > showThreshold) {
      this.setState({
        lastLoading: moment().unix(),
        status: "running",
      });
    }

    if (this.state.status === "running") {
      this.setState({
        display: "show",
      });
    }

    if (this.state.totalRequest == 0) {
      this.setState({
        display: "hide",
        status: "stop"
      });
    }

  }

  render() {
    if (this.state.display === "show") {
      return(<LoaderProgress variant="indeterminate" />);
    }

    return(null);
  }

}
export default withRouter(Loader);
